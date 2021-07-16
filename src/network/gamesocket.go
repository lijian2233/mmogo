package network

import (
	"errors"
	"fmt"
	linklist "github.com/emirpasic/gods/lists/singlylinkedlist"
	"mmogo/buffer"
	"mmogo/common/locker"
	"mmogo/common/logger"
	"mmogo/pack"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type TcpSocket struct {
	conn  net.Conn
	state uint32

	//peer ip port
	peerAddr string

	//本端ip port
	localAddr string

	//发送缓冲区
	sendBuf  *buffer.RingBuffer
	sendCond *sync.Cond

	//当发送缓冲区已满, 新发送的数据挂接到此列表
	sendBuffList *linklist.List

	//接受缓冲区
	recvLock locker.CASLock
	recvBuf  *buffer.RingBuffer

	exitSendChan chan bool
	exitRecvChan chan bool

	hasCloseSend bool
	hasCloseRecv bool

	remainBytes uint32

	//socket lock
	socketLock locker.CASLock

	handler func(*pack.WorldPacket)
	log     *logger.Logger
}

const max_buff_size = 0x400000 //4M
var Err_Conncet_Unknown = errors.New("unknown")
var Err_Max_Buff_Size = errors.New("buff size must less 4M")
var Err_Socket_Not_Open = errors.New("socket not open")
var Err_Conn_Is_Nil = errors.New("conn is nil")
var Err_Handler_Packet_Fn = errors.New("handler packet func must not nil")
var Err_Send_Packet_Is_Nil = errors.New("send ni packet")

func NewTcpSocket(addr string, port uint16, sendBuffSize, recvBuffSize uint32, handler func(*pack.WorldPacket)) (*TcpSocket, error) {
	if sendBuffSize > max_buff_size || recvBuffSize > max_buff_size {
		return nil, Err_Max_Buff_Size
	}

	if handler == nil {
		return nil, Err_Handler_Packet_Fn
	}

	socket := &TcpSocket{state: SOCKET_STATE_CREATE}
	socket.peerAddr = fmt.Sprintf("%s:%d", addr, port)
	socket.sendBuf = buffer.NewRingBuffer(int(sendBuffSize))
	socket.recvBuf = buffer.NewRingBuffer(int(recvBuffSize))
	socket.sendCond = sync.NewCond(&sync.Mutex{})
	socket.sendBuffList = linklist.New()
	socket.handler = handler
	return socket, nil
}

func NewConnSocket(conn net.Conn, sendBuffSize, recvBuffSize uint32, handler func(*pack.WorldPacket), log *logger.Logger) (*TcpSocket, error) {
	if sendBuffSize > max_buff_size || recvBuffSize > max_buff_size {
		return nil, Err_Max_Buff_Size
	}

	if conn == nil {
		return nil, Err_Conn_Is_Nil
	}

	if handler == nil {
		return nil, Err_Handler_Packet_Fn
	}

	socket := &TcpSocket{state: SOCKET_STATE_OPEN}
	socket.localAddr = conn.LocalAddr().String()
	socket.peerAddr = conn.RemoteAddr().String()

	socket.sendBuf = buffer.NewRingBuffer(int(sendBuffSize))
	socket.recvBuf = buffer.NewRingBuffer(int(recvBuffSize))
	socket.sendCond = sync.NewCond(&sync.Mutex{})
	socket.sendBuffList = linklist.New()
	socket.handler = handler
	socket.conn = conn

	socket.exitSendChan = make(chan bool, 1)
	socket.exitRecvChan = make(chan bool, 1)
	socket.log = log

	go socket.sendSocketData()
	go socket.recvMsg()

	return socket, nil
}

func (socket *TcpSocket) SetLog(log *logger.Logger) {
	socket.log = log
}

func (socket *TcpSocket) sendSocketData() {
	for {
		if atomic.LoadUint32(&socket.state) == SOCKET_STATE_CLOSE {
			socket.exitSendChan <- true
			break
		}

		socket.sendCond.L.Lock()
		if socket.sendBuffList.Empty() {
			if socket.log != nil {
				socket.log.Infof("socket %v :%v wait send data", unsafe.Pointer(socket), socket.localAddr)
			}
			socket.sendCond.Wait()
		}

		if socket.log != nil {
			socket.log.Infof("socket %v handle send data :%v", unsafe.Pointer(socket), socket.sendBuffList.Size())
		}
		for {
			if !socket.sendBuffList.Empty() {
				d, _ := socket.sendBuffList.Get(0)
				packet, _ := d.(*pack.WorldPacket)
				if int(packet.GetSize()) <= socket.sendBuf.Free() {
					socket.sendBuf.Write(packet.GetContent())
					socket.sendBuffList.Remove(0)
				}else{
					if socket.sendBuf.IsEmpty(){
						socket.sendBuffList.Remove(0)
						socket.sendCond.L.Unlock()
						_, err := socket.conn.Write(packet.GetContent())
						if err != nil && socket.log != nil {
							socket.log.Infof("socket :%v send data error %v", socket.localAddr, err)
						}
					}else{
						socket.sendCond.L.Unlock()
						_, err := socket.conn.Write(socket.sendBuf.UnsafeReadBytes())
						if err != nil && socket.log != nil {
							socket.log.Infof("socket :%v send data error %v", socket.localAddr, err)
						}
						socket.sendBuf.Adjust()
					}
					//回到主循环,主要用于判断socket退出
					break
				}
			}else{
				socket.sendCond.L.Unlock()
				if !socket.sendBuf.IsEmpty(){
					_, err := socket.conn.Write(socket.sendBuf.UnsafeReadBytes())
					if err != nil && socket.log != nil {
						socket.log.Infof("socket :%v send data error %v", socket.localAddr, err)
					}
					socket.sendBuf.Reset()
				}
				break
			}
		}
	}
}

func (socket *TcpSocket) onClose() {
	socket.conn.Close()
}

func (socket *TcpSocket) closeSend() {
	defer func() {
		recover()
	}()

	socket.socketLock.Lock()
	if socket.hasCloseSend {
		socket.socketLock.Unlock()
		return
	}
	socket.hasCloseSend = true
	socket.socketLock.Unlock()
	socket.sendCond.Signal()
}

func (socket *TcpSocket) closeRecv() {
	defer func() {
		recover()
	}()

	socket.socketLock.Lock()
	if socket.hasCloseRecv {
		socket.socketLock.Unlock()
		return
	}
	socket.hasCloseRecv = true
	socket.socketLock.Unlock()
}

func (socket *TcpSocket) Close() {
	for {
		old := atomic.LoadUint32(&socket.state)
		if old != SOCKET_STATE_CLOSE {
			if atomic.CompareAndSwapUint32(&socket.state, old, SOCKET_STATE_CLOSE) {
				if socket.exitSendChan != nil {
					socket.closeSend()
					socket.closeRecv()
					socket.onClose()
				}
			}
		} else {
			break
		}
	}
}

func (socket *TcpSocket) IsOpen() bool {
	return atomic.LoadUint32(&socket.state) == SOCKET_STATE_OPEN
}

func (socket *TcpSocket) Connect(timeout time.Duration) error {
	if atomic.CompareAndSwapUint32(&socket.state, SOCKET_STATE_CREATE, SOCKET_STATE_CONNECTING) {
		conn, err := net.DialTimeout("tcp", socket.peerAddr, timeout)
		if err != nil {
			return err
		}
		if atomic.CompareAndSwapUint32(&socket.state, SOCKET_STATE_CONNECTING, SOCKET_STATE_OPEN) {
			socket.conn = conn
			socket.localAddr = conn.LocalAddr().String()
		} else {
			return Err_Conncet_Unknown
		}
	}
	return Err_Conncet_Unknown
}

func (socket *TcpSocket) SendPacket(packet *pack.WorldPacket) error {
	if !socket.IsOpen() {
		return Err_Socket_Not_Open
	}

	if packet == nil {
		if socket.log != nil {
			socket.log.Errorf("send ni packet %s", string(debug.Stack()))
		}
		return Err_Send_Packet_Is_Nil
	}

	socket.sendCond.L.Lock()
	defer socket.sendCond.L.Unlock()

	f := socket.sendBuffList.Empty()
	socket.sendBuffList.Add(packet)

	if f {
		socket.sendCond.Signal()
	}
	return nil
}

func (socket *TcpSocket) recvMsg() {
	for {
		if atomic.LoadUint32(&socket.state) == SOCKET_STATE_CLOSE {
			socket.exitRecvChan <- true
			break
		}

		bytes := socket.recvBuf.UnsafeWriteSpace()
		if bytes == nil {
			//缓冲区满了
			socket.Close()
			return
		}
		if uint32(len(bytes)) < pack.GetWorldPacketHeaderSize() {
			if socket.recvBuf.Free() > int(pack.GetWorldPacketHeaderSize()) {
				socket.recvBuf.Adjust()
				bytes = socket.recvBuf.UnsafeWriteSpace()
				if len(bytes) < int(pack.GetWorldPacketHeaderSize()) {
					socket.Close()
					return
				}
			} else {
				//缓冲区满了
				socket.Close()
				return
			}
		}

		n, err := socket.conn.Read(bytes)
		if err != nil {
			//记录日志
			socket.Close()
		}
		socket.recvBuf.IncWriteIndex(n)

		for {
			if socket.recvBuf.Length() < int(pack.GetWorldPacketHeaderSize()) {
				break
			}

			rbuf := socket.recvBuf.UnsafeReadBytes()
			_, bodySize, op, ok := pack.ParsePacketHeader(rbuf)
			if ok {
				if socket.recvBuf.Length() >= int(bodySize) {
					if len(rbuf) >= int(bodySize) {
						//full packet
						packet, _ := pack.BuildWorldPacket(rbuf, false)
						socket.recvBuf.Erase(int(bodySize))
						socket.handler(packet)
					} else {
						worldPack := pack.NewWorldPacket(op, bodySize)
						contentPtr := uintptr(unsafe.Pointer(&rbuf[0])) + uintptr(pack.GetWorldPacketHeaderSize())
						h := [3]uintptr{contentPtr, uintptr(uint32(len(rbuf)) - pack.GetWorldPacketHeaderSize()),
							uintptr(uint32(len(rbuf)) - pack.GetWorldPacketHeaderSize())}
						buf := *(*[]byte)(unsafe.Pointer(&h))
						worldPack.WriteBytes(buf, len(buf))

						socket.recvBuf.Erase(len(rbuf))
						left := int(bodySize) - len(rbuf)
						rbuf = socket.recvBuf.UnsafeReadBytes()
						if len(rbuf) < left {
							panic("recv buf")
						}
						worldPack.WriteBytes(rbuf, left)
						socket.recvBuf.Erase(left)
					}
				}else{
					break
				}
			}else{
				if socket.log != nil{
					socket.log.Info("socket remote addr %v recv error format packet, will close", socket.peerAddr)
				}
				break
			}
		}
	}
}
