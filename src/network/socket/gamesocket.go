package socket

import (
	"fmt"
	"mmogo/interface"
	"mmogo/lib/buffer"
	"mmogo/lib/locker"
	"mmogo/lib/packet"
	"mmogo/lib/queue"
	"mmogo/network"
	"net"
	"runtime/debug"
	"sync/atomic"
	"time"
)

type GameSocket struct {
	tcpSocket
	connTimeOut time.Duration

	//发送缓冲区
	sendBuf   *buffer.RingBuffer
	sendQueue *queue.UnsafeFastQueue

	//接受缓冲区
	recvLock locker.CASLock
	recvBuf  *buffer.RingBuffer

	exitSendChan chan bool
	exitRecvChan chan bool

	hasCloseSend bool
	hasCloseRecv bool

	//socket lock
	socketLock locker.CASLock
	packetType _interface.BinaryPacket

	handleBinaryPacket func(_interface.BinaryPacket)
	logger             _interface.Logger
	mgr                *GameSocketMgr
}

const max_buff_size = 0x400000 //4M

type GameSocketOpt func(socket *GameSocket)

func WithGameLog(logger _interface.Logger) GameSocketOpt {
	return func(socket *GameSocket) {
		socket.logger = logger
	}
}

func WithSocketMgr(mgr *GameSocketMgr) GameSocketOpt {
	return func(socket *GameSocket) {
		socket.mgr = mgr
	}
}

func WithGameHandlePacket(h func(packet _interface.BinaryPacket)) GameSocketOpt {
	return func(socket *GameSocket) {
		socket.handleBinaryPacket = h
	}
}

func WithGameSendBuffSize(size int) GameSocketOpt {
	return func(socket *GameSocket) {
		if size > max_buff_size {
			size = max_buff_size
		}
		socket.sendBuf = buffer.NewRingBuffer(size)
	}
}

func WithGameRecivBuffSize(size int) GameSocketOpt {
	return func(socket *GameSocket) {
		if size > max_buff_size {
			size = max_buff_size
		}

		socket.recvBuf = buffer.NewRingBuffer(size)
	}
}

func WithGameConnectTimeOut(timeOut time.Duration) GameSocketOpt {
	return func(socket *GameSocket) {
		if timeOut > 0 {
			socket.connTimeOut = timeOut
		}
	}
}

func WithGameBinaryPacket(packet _interface.BinaryPacket) GameSocketOpt {
	return func(socket *GameSocket) {
		socket.packetType = packet
	}
}

func (socket *GameSocket) init() {
	socket.sendQueue = queue.NewUnsafeFastQueue()

	if socket.sendBuf == nil {
		socket.sendBuf = buffer.NewRingBuffer(64 * 1024)
	}

	if socket.recvBuf == nil {
		socket.recvBuf = buffer.NewRingBuffer(64 * 1024)
	}

	socket.exitSendChan = make(chan bool, 1)
	socket.exitRecvChan = make(chan bool, 1)

	if socket.packetType == nil {
		socket.packetType = &packet.WorldPacket{}
	}

	socket.localAddr = socket.conn.LocalAddr().String()
	socket.peerAddr = socket.conn.RemoteAddr().String()

	if socket.logger == nil {
		socket.logger = defualtLogger(fmt.Sprintf("%s:%s", socket.peerAddr, socket.localAddr))
	}

	go socket.routineSendMsg()
	go socket.routineRecvMsg()
}

func NewGameSocket(addr string, port uint16, opts ...GameSocketOpt) (*GameSocket, error) {
	socket := &GameSocket{}
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

	if socket.mgr != nil {
		socket.mgr.AddSocket(socket)
	}
	socket.init()

	return socket, nil
}

func NewConnSocket(conn net.Conn, opts ...GameSocketOpt) (*GameSocket, error) {
	if conn == nil {
		return nil, Err_Conn_Is_Nil
	}

	socket := &GameSocket{}
	socket.state = network.SOCKET_STATE_OPEN
	for _, fn := range (opts) {
		fn(socket)
	}

	socket.conn = conn
	socket.init()
	return socket, nil
}

func (socket *GameSocket) routineSendMsg() {
	time.Sleep(time.Second * 2)
	for {
		if atomic.LoadUint32(&socket.state) == network.SOCKET_STATE_CLOSE {
			socket.exitSendChan <- true
			break
		}

		buffList := socket.sendQueue.PopALl()
		if buffList.Empty(){
			continue
		}

		totalPacketSize := 0
		for ; !buffList.Empty(); {
			d, _ := buffList.Get(0)
			packet, _ := d.(_interface.BinaryPacket)
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
					socket.logger.Infof("socket :%v send data error %v", socket.localAddr, err)

				} else {
					_, err := socket.conn.Write(socket.sendBuf.UnsafeReadBytes())
					socket.logger.Infof("socket :%v send data error %v", socket.localAddr, err)
					socket.sendBuf.Reset()
				}
			}
		}
		if !socket.sendBuf.IsEmpty() {
			//直接发送packet
			_, err := socket.conn.Write(socket.sendBuf.UnsafeReadBytes())
			socket.logger.Infof("socket :%v send data error %v", socket.localAddr, err)
			socket.sendBuf.Reset()
		}
	}
}

func (socket *GameSocket) onClose() {
	socket.conn.Close()
	if socket.mgr != nil {
		socket.mgr.RemoveSocket(socket)
	}
}

func (socket *GameSocket) closeSend() {
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

	socket.sendQueue.Signal()
}

func (socket *GameSocket) closeRecv() {
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

func (socket *GameSocket) Close() {
	for {
		old := atomic.LoadUint32(&socket.state)
		if old != network.SOCKET_STATE_CLOSE {
			if atomic.CompareAndSwapUint32(&socket.state, old, network.SOCKET_STATE_CLOSE) {
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

func (socket *GameSocket) IsOpen() bool {
	return atomic.LoadUint32(&socket.state) == network.SOCKET_STATE_OPEN
}

func (socket *GameSocket) Connect(timeout time.Duration) error {
	if atomic.CompareAndSwapUint32(&socket.state, network.SOCKET_STATE_CREATE, network.SOCKET_STATE_CONNECTING) {
		conn, err := net.DialTimeout("tcp", socket.peerAddr, timeout)
		if err != nil {
			return err
		}
		if atomic.CompareAndSwapUint32(&socket.state, network.SOCKET_STATE_CONNECTING, network.SOCKET_STATE_OPEN) {
			socket.conn = conn
			socket.localAddr = conn.LocalAddr().String()
		} else {
			return Err_Conncet_Unknown
		}
	}
	return Err_Conncet_Unknown
}

func (socket *GameSocket) SendPacket(packet _interface.BinaryPacket) error {
	if !socket.IsOpen() {
		return Err_Socket_Not_Open
	}

	if packet == nil {
		socket.logger.Errorf("send ni packet %s", string(debug.Stack()))
		return Err_Send_Packet_Is_Nil
	}

	socket.sendQueue.Add(packet)

	return nil
}

func (socket *GameSocket) routineRecvMsg() {
	for {
		if atomic.LoadUint32(&socket.state) == network.SOCKET_STATE_CLOSE {
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
				socket.logger.Info("socket remote addr %v recv error format packet, will close", socket.peerAddr)
				break
			}
		}
	}
}
