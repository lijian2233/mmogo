package socket

import (
	"errors"
	"fmt"
	"mmogo/network"
	"net"
	"sync/atomic"
	"time"
)

var err_no_connect = errors.New("error socket not connected")

type tcpSocket struct {
	conn  net.Conn
	state uint32

	peerAddr  string //peer ip port
	localAddr string //本端ip port
}

func (g *tcpSocket) acceptConn(conn net.Conn) {
	g.conn = conn
	atomic.StoreUint32(&g.state, network.SOCKET_STATE_OPEN)
}

func (g *tcpSocket) connect(addr string, port int16, timeout time.Duration) error {
	connectAddr := fmt.Sprintf("%s:%d", addr, port)
	if atomic.CompareAndSwapUint32(&g.state, network.SOCKET_STATE_CREATE, network.SOCKET_STATE_CONNECTING) {
		conn, err := net.DialTimeout("tcp", connectAddr, timeout)
		if err != nil {
			return err
		}
		if atomic.CompareAndSwapUint32(&g.state, network.SOCKET_STATE_CONNECTING, network.SOCKET_STATE_OPEN) {
			g.conn = conn
		} else {
			return Err_Conncet_Unknown
		}
	}
	return Err_Conncet_Unknown
}

func (g *tcpSocket) sendSync(msg []byte, timeout ...time.Duration) (int, error) {
	if g.conn == nil {
		return 0, err_no_connect
	}

	if len(timeout) > 0 {
		g.conn.SetWriteDeadline(time.Now().Add(timeout[0]))
	}
	return g.conn.Write(msg)
}

func (g *tcpSocket) readSync(b []byte, timeout ...time.Duration) (int, error) {
	if g.conn == nil {
		return 0, err_no_connect
	}

	if len(timeout) > 0 {
		g.conn.SetReadDeadline(time.Now().Add(timeout[0]))
	}
	return g.conn.Read(b)
}

func (g *tcpSocket) closeSync() error {
	err := g.conn.Close()
	if err != nil {
		return err
	}
	atomic.StoreUint32(&g.state, network.SOCKET_STATE_CLOSE)
	return nil
}
