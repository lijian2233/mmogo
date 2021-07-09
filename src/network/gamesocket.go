package network

import (
	"errors"
	"fmt"
	linklist "github.com/emirpasic/gods/lists/singlylinkedlist"
	"mmogo/buffer"
	"mmogo/common/locker"
	"mmogo/pack"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type TcpSocket struct {
	conn  net.Conn
	state uint32

	//peer ip port
	peerAddr string

	//本端ip port
	localAddr string

	//发送缓冲区
	sendLock locker.CASLock
	sendBuf  *buffer.RingBuffer
	sendCond *sync.Cond

	//当发送缓冲区已满, 新发送的数据挂接到此列表
	sendBuffList *linklist.List

	//接受缓冲区
	recvLock locker.CASLock
	recvBuf  *buffer.RingBuffer
	recvCond *sync.Cond

	exitSendChan chan bool
	exitRecvChan chan bool

	hasCloseSend bool
	hasCloseRecv bool

	remainBytes uint32

	//socket lock
	socketLock locker.CASLock
}

const max_buff_size = 0x400000 //4M
var Err_Conncet_Unknown = errors.New("unknown")
var Err_Max_Buff_Size = errors.New("buff size must less 4M")
var Err_Socket_Not_Open = errors.New("socket not open")
var Err_Conn_Is_Nil = errors.New("conn is nil")

func NewTcpSocket(addr string, port uint16, sendBuffSize, recvBuffSize uint32) (*TcpSocket, error) {
	if sendBuffSize > max_buff_size || recvBuffSize > max_buff_size {
		return nil, Err_Max_Buff_Size
	}
	socket := &TcpSocket{state: SOCKET_STATE_CREATE}
	socket.peerAddr = fmt.Sprintf("%s:%d", addr, port)
	socket.sendBuf = buffer.NewRingBuffer(int(sendBuffSize))
	socket.recvBuf = buffer.NewRingBuffer(int(recvBuffSize))
	socket.sendCond = sync.NewCond(&socket.sendLock)
	socket.recvCond = sync.NewCond(&socket.recvLock)
	socket.sendBuffList = linklist.New()
	return socket, nil
}

func NewConnSocket(conn net.Conn, sendBuffSize, recvBuffSize uint32) (*TcpSocket, error) {
	if sendBuffSize > max_buff_size || recvBuffSize > max_buff_size {
		return nil, Err_Max_Buff_Size
	}

	if conn == nil {
		return nil, Err_Conn_Is_Nil
	}

	socket := &TcpSocket{state: SOCKET_STATE_OPEN}
	socket.localAddr = conn.LocalAddr().String()
	socket.peerAddr = conn.RemoteAddr().String()

	socket.sendBuf = buffer.NewRingBuffer(int(sendBuffSize))
	socket.recvBuf = buffer.NewRingBuffer(int(recvBuffSize))
	socket.sendCond = sync.NewCond(&socket.sendLock)
	socket.recvCond = sync.NewCond(&socket.recvLock)
	socket.sendBuffList = linklist.New()

	socket.exitSendChan = make(chan bool, 1)
	socket.exitRecvChan = make(chan bool, 1)
	go socket.sendSocketData()
	return socket, nil
}

func (socket *TcpSocket) sendSocketData() {
	for {
		if atomic.LoadUint32(&socket.state) == SOCKET_STATE_CLOSE{
			socket.exitSendChan <- true
			break
		}

		socket.sendLock.Lock()
		if socket.sendBuf.IsEmpty() && socket.sendBuffList.Empty() {
			socket.sendCond.Wait()
		}
		socket.sendLock.Unlock()

		for {
			if socket.sendBuf.IsEmpty() && socket.sendBuffList.Empty() {
				break
			}

			socket.sendLock.Lock()
			if !socket.sendBuf.IsEmpty() {
				bytes := socket.sendBuf.UnsafeBytes()
				socket.sendLock.Unlock()
				_, err := socket.conn.Write(bytes)
				if err != nil {
					socket.Close()
					return
				}
			} else {
				if !socket.sendBuffList.Empty() {
					d, _ := socket.sendBuffList.Get(0)
					bytes, _ := d.([]byte)
					socket.sendLock.Unlock()

					_, err := socket.conn.Write(bytes)
					if err != nil {
						socket.Close()
					}
				}
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

func (socket *TcpSocket) SendMsg(bytes []byte) error {
	if socket.IsOpen() {
		return Err_Socket_Not_Open
	}

	socket.sendLock.Lock()
	defer socket.sendLock.Unlock()

	if !socket.sendBuffList.Empty() {
		socket.sendBuffList.Add(bytes)
		return nil
	}

	if len(bytes) > socket.sendBuf.GetWriteSpace() {
		socket.sendBuf.Write(bytes)
	} else {
		socket.sendBuffList.Add(bytes)
	}
	return nil
}

func (socket *TcpSocket) recvMsg() {
	for{
		if atomic.LoadUint32(&socket.state) ==  SOCKET_STATE_CLOSE{
			socket.exitRecvChan <- true
			break
		}

		bytes := socket.recvBuf.UnsafeBytes()
		if bytes == nil || uint32(len(bytes)) <pack.GetWorldPacketHeaderSize(){
			//缓冲区满了
			socket.Close()
			return
		}

		n, err := socket.conn.Read(bytes)
		if err != nil{
			//记录日志
			socket.Close()
		}

		if uint32(n) < pack.GetWorldPacketHeaderSize(){
			continue
		}



	}
}
