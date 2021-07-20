package network

import (
	"errors"
	"fmt"
	linklist "github.com/emirpasic/gods/lists/singlylinkedlist"
	"mmogo/buffer"
	"mmogo/common/locker"
	"mmogo/gameInterface"
	"mmogo/pack"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type TcpSocket struct {
	connTimeOut int
	sendTimeOut int
	recvTimeOut int

	conn  net.Conn
	state uint32

	//peer ip port
	peerAddr string

	//本端ip port
	localAddr string

	//发送缓冲区
	sendBuf  *buffer.RingBuffer
	sendCond *sync.Cond

	//双写缓冲,避免阻塞发送携程
	sendBuffList  *linklist.List
	sendBuffList1 *linklist.List
	sendBuffList2 *linklist.List

	//接受缓冲区
	recvLock locker.CASLock
	recvBuf  *buffer.RingBuffer

	exitSendChan chan bool
	exitRecvChan chan bool

	hasCloseSend bool
	hasCloseRecv bool

	//socket lock
	socketLock locker.CASLock
	packetType gameInterface.BinaryPacket

	handleBinaryPacket func(gameInterface.BinaryPacket)
	log                gameInterface.Log
}

const max_buff_size = 0x400000 //4M
var Err_Conncet_Unknown = errors.New("unknown")
var Err_Max_Buff_Size = errors.New("buff size must less 4M")
var Err_Socket_Not_Open = errors.New("socket not open")
var Err_Conn_Is_Nil = errors.New("conn is nil")
var Err_Handler_Packet_Fn = errors.New("handler packet func must not nil")
var Err_Send_Packet_Is_Nil = errors.New("send ni packet")

type Opt func(socket *TcpSocket)

func WithLog(log gameInterface.Log) Opt {
	return func(socket *TcpSocket) {
		socket.log = log
	}
}

func WithHandlePacket(h func(packet gameInterface.BinaryPacket)) Opt {
	return func(socket *TcpSocket) {
		socket.handleBinaryPacket = h
	}
}

func WithSendBuffSize(size int) Opt {
	return func(socket *TcpSocket) {
		if size > max_buff_size {
			size = max_buff_size
		}
		socket.sendBuf = buffer.NewRingBuffer(size)
	}
}

func WithRecivBuffSize(size int) Opt {
	return func(socket *TcpSocket) {
		if size > max_buff_size {
			size = max_buff_size
		}

		socket.recvBuf = buffer.NewRingBuffer(size)
	}
}

func WithConnectTimeOut(timeOut int) Opt {
	return func(socket *TcpSocket) {
		if timeOut > 0 {
			socket.connTimeOut = timeOut
		}
	}
}

func WithSendTimeOut(timeOut int) Opt {
	return func(socket *TcpSocket) {
		if timeOut > 0 {
			socket.sendTimeOut = timeOut
		}
	}
}

func WithRecvTimeOut(timeOut int) Opt {
	return func(socket *TcpSocket) {
		if timeOut > 0 {
			socket.recvTimeOut = timeOut
		}
	}
}

func WithBinaryPacket(packet gameInterface.BinaryPacket) Opt {
	return func(socket *TcpSocket) {
		socket.packetType = packet
	}
}

func (socket *TcpSocket) initBuffer() {
	socket.sendCond = sync.NewCond(&sync.Mutex{})
	socket.sendBuffList1 = linklist.New()
	socket.sendBuffList2 = linklist.New()
	socket.sendBuffList = socket.sendBuffList1
	if socket.sendBuf == nil {
		socket.sendBuf = buffer.NewRingBuffer(64 * 1024)
	}

	if socket.recvBuf == nil {
		socket.recvBuf = buffer.NewRingBuffer(64 * 1024)
	}
}

func NewTcpSocket(addr string, port uint16, opts ...Opt) (*TcpSocket, error) {
	socket := &TcpSocket{state: SOCKET_STATE_CREATE}
	for _, fn := range (opts) {
		fn(socket)
	}

	socket.peerAddr = fmt.Sprintf("%s:%d", addr, port)

	var err error
	if socket.connTimeOut == 0 {
		socket.conn, err = net.Dial("tcp", socket.peerAddr)
	} else {
		socket.conn, err = net.DialTimeout("tcp", socket.peerAddr, time.Duration(socket.connTimeOut))
	}
	if err != nil {
		return nil, err
	}

	socket.initBuffer()

	if socket.packetType == nil {
		socket.packetType = &pack.WorldPacket{}
	}
	socket.exitSendChan = make(chan bool, 1)
	socket.exitRecvChan = make(chan bool, 1)
	go socket.sendSocketData()
	go socket.recvMsg()
	return socket, nil
}

func (socket *TcpSocket) SetHandler(fn func(packet gameInterface.BinaryPacket)) {
	socket.handleBinaryPacket = fn
}

func (socket *TcpSocket) SetLog(log gameInterface.Log) {
	socket.log = log
}

func NewConnSocket(conn net.Conn, opts ...Opt) (*TcpSocket, error) {
	if conn == nil {
		return nil, Err_Conn_Is_Nil
	}

	socket := &TcpSocket{state: SOCKET_STATE_OPEN}
	for _, fn := range (opts) {
		fn(socket)
	}

	if conn == nil {
		return nil, Err_Conn_Is_Nil
	}

	socket.localAddr = conn.LocalAddr().String()
	socket.peerAddr = conn.RemoteAddr().String()
	socket.conn = conn
	socket.initBuffer()
	socket.exitSendChan = make(chan bool, 1)
	socket.exitRecvChan = make(chan bool, 1)

	if socket.packetType == nil {
		socket.packetType = &pack.WorldPacket{}
	}

	go socket.sendSocketData()
	go socket.recvMsg()

	return socket, nil
}

func (socket *TcpSocket) sendSocketData() {
	time.Sleep(time.Second * 2)
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

		buffList := socket.sendBuffList
		if socket.sendBuffList == socket.sendBuffList1 {
			socket.sendBuffList = socket.sendBuffList2
		} else {
			socket.sendBuffList = socket.sendBuffList1
		}
		socket.sendCond.L.Unlock()

		totalPacketSize := 0
		for ; !buffList.Empty(); {
			d, _ := buffList.Get(0)
			packet, _ := d.(gameInterface.BinaryPacket)
			totalPacketSize += int(packet.PacketSize())
			if int(packet.PacketSize()) <= socket.sendBuf.Free() {
				//优先写到缓冲区
				socket.sendBuf.Write(packet.GetPacket(false))
				buffList.Remove(0)
			} else {
				if socket.sendBuf.IsEmpty() {
					buffList.Remove(0)
					//直接发送packet
					_, err := socket.conn.Write(packet.GetPacket(false))
					if err != nil && socket.log != nil {
						socket.log.Infof("socket :%v send data error %v", socket.localAddr, err)
					}
				} else {
					_, err := socket.conn.Write(socket.sendBuf.UnsafeReadBytes())
					if err != nil && socket.log != nil {
						socket.log.Infof("socket :%v send data error %v", socket.localAddr, err)
					}
					socket.sendBuf.Reset()
				}
			}
		}
		if !socket.sendBuf.IsEmpty(){
			//直接发送packet
			_, err := socket.conn.Write(socket.sendBuf.UnsafeReadBytes())
			if err != nil && socket.log != nil {
				socket.log.Infof("socket :%v send data error %v", socket.localAddr, err)
			}
			socket.sendBuf.Reset()
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

func (socket *TcpSocket) SendPacket(packet gameInterface.BinaryPacket) error {
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

		headerSize := int(socket.packetType.HeaderSize())
		if len(bytes) < headerSize {
			if socket.recvBuf.Free() > headerSize {
				socket.recvBuf.Adjust()
				bytes = socket.recvBuf.UnsafeWriteSpace()
				if len(bytes) < headerSize {
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
			if socket.recvBuf.Length() < headerSize {
				break
			}

			rbuf := socket.recvBuf.UnsafeReadBytes()
			packetSize, ok := socket.packetType.ParsePacketHeader(rbuf)
			if ok {
				if socket.recvBuf.Length() >= int(packetSize) {
					if len(rbuf) >= int(packetSize) {
						//full packet
						packet, _ := socket.packetType.BuildPacket(rbuf, false)
						socket.recvBuf.Erase(int(packetSize))
						socket.handleBinaryPacket(packet)
					} else {
						buf := make([]byte, packetSize, packetSize)
						copy(buf, rbuf)
						wlen := len(rbuf)
						socket.recvBuf.Erase(len(rbuf))
						left := int(packetSize) - len(rbuf)
						rbuf = socket.recvBuf.UnsafeReadBytes()
						if len(rbuf) < left {
							panic("recv buf")
						}
						copy(buf[wlen:], rbuf)
						socket.recvBuf.Erase(left)
						packet, _ := socket.packetType.BuildPacket(rbuf, true)
						socket.handleBinaryPacket(packet)
					}
				} else {
					break
				}
			} else {
				if socket.log != nil {
					socket.log.Info("socket remote addr %v recv error format packet, will close", socket.peerAddr)
				}
				break
			}
		}
	}
}
