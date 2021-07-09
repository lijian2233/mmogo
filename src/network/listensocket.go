package network

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
)

var Handle_Not_Set_Err = errors.New("handle not set error")


type ListenSocket struct {
	addr     string
	port     uint16
	state    uint32
	listener *net.TCPListener
	handler  func(conn net.Conn) ()
}

func NewListenSocket(addr string, port uint16, handler func(conn net.Conn) ()) (*ListenSocket, error) {
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

	return &ListenSocket{addr: addr, port: port, listener: tcpListen, state:SOCKET_STATE_OPEN, handler:handler}, nil
}

func (l *ListenSocket) Start() {
	for {
		//等待客户的连接
		conn, err3 := l.listener.Accept();
		//如果有错误直接跳过
		if err3 != nil {
			continue
		}
		l.handler(conn)
	}
}

func (l *ListenSocket) Stop() {
	if atomic.CompareAndSwapUint32(&l.state, SOCKET_STATE_OPEN, SOCKET_STATE_CLOSE) {
		l.listener.Close()
	}
}

func (l *ListenSocket) IsClose() bool {
	return l.state == SOCKET_STATE_CLOSE
}

