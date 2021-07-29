package listen

import (
	"errors"
	"fmt"
	_interface "mmogo/interface"
	"mmogo/network"
	"net"
	"sync/atomic"
)

var Handle_Not_Set_Err = errors.New("handle not set error")

type acceptFn func(conn net.Conn)

type ListenSocket struct {
	addr     string
	port     uint16
	state    uint32
	closed   bool
	logger   _interface.Logger
	listener *net.TCPListener
	handler  acceptFn
	closeCh  chan bool
}

type Opt func(l *ListenSocket)

func WithLog(logger _interface.Logger) Opt {
	return func(l *ListenSocket) {
		l.logger = logger
	}
}

func WithHandler(fn func(conn net.Conn)) Opt {
	return func(l *ListenSocket) {
		l.handler = fn
	}
}

func NewListenSocket(addr string, port uint16, opts ...Opt) (*ListenSocket, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}

	tcpListen, err2 := net.ListenTCP("tcp", tcpAddr)
	if err2 != nil {
		return nil, err2
	}

	l := &ListenSocket{addr: addr, port: port, listener: tcpListen, state: network.SOCKET_STATE_OPEN}

	if len(opts) > 0 {
		for _, opt := range opts {
			opt(l)
		}
	}
	if l.logger == nil {
		l.logger = defualtLogger(fmt.Sprintf("%s:%d", addr, port))
	}

	if l.handler == nil {
		l.handler = defaultHandler()
	}
	return l, nil
}

func (l *ListenSocket) Start() {
	if !atomic.CompareAndSwapUint32(&l.state, network.SOCKET_STATE_OPEN, network.SOCKET_STATE_LISTEN) {
		l.logger.Errorf("listen socket :%v start error:%v", l.port, "not open state")
		return
	}
	for {
		//等待客户的连接
		conn, err3 := l.listener.Accept();
		//如果有错误直接跳过
		if err3 != nil {
			if atomic.LoadUint32(&l.state) == network.SOCKET_STATE_CLOSING {
				atomic.StoreUint32(&l.state, network.SOCKET_STATE_CLOSE)
				l.closed = true
				l.closeCh <- true
				l.logger.Warningf("listen socket :%v exit", l.port)
				return
			}
			continue
		}
		l.handler(conn)
	}
}

func (l *ListenSocket) Stop() <- chan bool{
	s := atomic.LoadUint32(&l.state)

	switch s {
	case network.SOCKET_STATE_OPEN:
		{
			if atomic.CompareAndSwapUint32(&l.state, network.SOCKET_STATE_OPEN, network.SOCKET_STATE_CLOSE) {
				l.listener.Close()
				l.closed = true
				l.closeCh <- true
				return l.closeCh
			}
		}

	case network.SOCKET_STATE_LISTEN:
		{
			if atomic.CompareAndSwapUint32(&l.state, network.SOCKET_STATE_LISTEN, network.SOCKET_STATE_CLOSING) {
				l.listener.Close()
			}
		}
	}

	return l.closeCh
}

func (l *ListenSocket) IsClosed() bool {
	return l.closed
}
