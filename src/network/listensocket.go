package network

import (
	"errors"
	"fmt"
	"mmogo/common/logger"
	"net"
	"sync/atomic"
)

var Handle_Not_Set_Err = errors.New("handle not set error")

type ListenSocket struct {
	addr     string
	port     uint16
	state    uint32
	closed   bool
	log      logger.Logger
	listener *net.TCPListener
	handler  func(conn net.Conn) ()
}

func NewListenSocket(addr string, port uint16, log logger.Logger, handler func(conn net.Conn) ()) (*ListenSocket, error) {
	if handler == nil {
		return nil, Handle_Not_Set_Err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}

	tcpListen, err2 := net.ListenTCP("tcp", tcpAddr)
	if err2 != nil {
		return nil, err2
	}

	return &ListenSocket{addr: addr, port: port, listener: tcpListen, state: SOCKET_STATE_OPEN, handler: handler, log:log}, nil
}

func (l *ListenSocket) Start() {
	if !atomic.CompareAndSwapUint32(&l.state, SOCKET_STATE_OPEN, SOCKET_STATE_LISTEN) {
		l.log.Errorf("listen socket :%v start error:%v", l.port, "not open state")
		return
	}
	for {
		//等待客户的连接
		conn, err3 := l.listener.Accept();
		//如果有错误直接跳过
		if err3 != nil {
			if atomic.LoadUint32(&l.state) == SOCKET_STATE_CLOSING {
				atomic.StoreUint32(&l.state, SOCKET_STATE_CLOSE)
				l.closed = true
				l.log.Warningf("listen socket :%v exit", l.port)
				return
			}
			continue
		}
		l.handler(conn)
	}
}

func (l *ListenSocket) Stop() {
	s := atomic.LoadUint32(&l.state)

	switch s {
	case SOCKET_STATE_OPEN:
		{
			if atomic.CompareAndSwapUint32(&l.state, SOCKET_STATE_OPEN, SOCKET_STATE_CLOSE) {
				l.listener.Close()
				l.closed = true
			}
		}

	case SOCKET_STATE_LISTEN:
		{
			if atomic.CompareAndSwapUint32(&l.state, SOCKET_STATE_LISTEN, SOCKET_STATE_CLOSING){
				l.listener.Close()
			}
		}
	}
}

func (l *ListenSocket) IsClosed() bool {
	return l.closed
}
