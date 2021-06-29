package network

import (
	"fmt"
	"net"
	"time"
)

type TcpSocket struct {
	conn     net.Conn
	state    uint32
	peerAddr string
	perrPort uint16
}

func NewTcpSocket(addr string, port uint16) *TcpSocket {
	socket := &TcpSocket{peerAddr: addr, perrPort: port, state: SOCKET_STATE_CREATE}
	return socket
}

func (socket *TcpSocket) Connect(timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", socket.peerAddr, socket.perrPort), timeout)
	if err != nil {
		return err
	}

	socket.conn = conn
	return nil
}

func (socket *TcpSocket) OnConnect() {
}

func (socket *TcpSocket) SendMsg(bytes []byte) (n int, err error) {
	return socket.conn.Write(bytes)
}

func (socket* TcpSocket) ReadMsg(bytes []byte) (n int, err error) {
	return socket.conn.Read(bytes)
}
