package socket

import (
	"bufio"
	"fmt"
	_interface "mmogo/interface"
	"mmogo/network"
	"net"
	"strings"
	"sync/atomic"
)

type TelnetSocket struct {
	tcpSocket
	logger _interface.Logger
}

type TelnetSocketOpt func(t *TelnetSocket)

func WithTelnetLog(logger _interface.Logger) TelnetSocketOpt {
	return func(t *TelnetSocket) {
		t.logger = logger
	}
}

func NewTelnetSocket(conn net.Conn, opts ...TelnetSocketOpt) (*TelnetSocket, error) {
	t := TelnetSocket{}
	t.conn = conn
	atomic.StoreUint32(&t.state, network.SOCKET_STATE_OPEN)

	if len(opts) > 0 {
		for _, opt := range (opts) {
			opt(&t)
		}
	}

	t.peerAddr = conn.RemoteAddr().String()
	t.localAddr = conn.LocalAddr().String()

	if t.logger == nil {
		t.logger = defualtLogger(fmt.Sprintf("telnet:%s:%s", t.peerAddr, t.localAddr))
	}
	go t.recvMsgRoutine()
	return &t, nil
}

func (t *TelnetSocket) Stop() {
	t.Close()
}

func (t *TelnetSocket) Close() {
	for {
		val := atomic.LoadUint32(&t.state)
		if val == network.SOCKET_STATE_CLOSE || val == network.SOCKET_STATE_CLOSING {
			return
		}

		if atomic.CompareAndSwapUint32(&t.state, val, network.SOCKET_STATE_CLOSE) {
			t.conn.Close()
		}
	}
}

func (t *TelnetSocket) recvMsgRoutine() {
	reader := bufio.NewReader(t.conn)

	for {
		str, err := reader.ReadString('\n')

		// telnet命令
		if err == nil {
			str = strings.TrimSpace(str)
			if !t.processTelnetCommand(str) {
				t.Close()
				break
			}
		} else {
			// 发生错误
			fmt.Println("Session closed")
			t.Close()
			break
		}
	}
}

func (t *TelnetSocket) processTelnetCommand(str string) bool {
	// @close指令表示终止本次会话
	if strings.HasPrefix(str, "@close") {
		t.logger.Infof("Session closed")
		// 告知外部需要断开连接
		return false
		// @shutdown指令表示终止服务进程
	} else if strings.HasPrefix(str, "@shutdown") {
		t.logger.Infof("Server shutdown")
		// 往通道中写入0, 阻塞等待接收方处理
		return false
	}

	// 打印输入的字符串
	return t.onCommand(str)
}

func (t *TelnetSocket) onCommand(cmd string) bool {
	t.logger.Infof("recv command %s ", cmd)
	t.conn.Write([]byte("reply" +  cmd + "\r\n"))
	return true
}
