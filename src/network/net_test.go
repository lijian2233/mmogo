package network

import (
	"fmt"
	"net"
	"testing"
	"time"
)


func handleAccept(conn net.Conn) {
	readgo(conn)
	time.Sleep(time.Second * 2)
	writego(conn)
	writego(conn)
}

func TestNewListenSocket(t *testing.T) {

	l, err := NewListenSocket("127.0.0.1", 8080, handleAccept)
	if err != nil {
		fmt.Println(err)
		return
	}

	go l.Start()

	c, err := clinetConncet("127.0.0.1", 8080)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(c)
	time.Sleep(time.Second * 50)
}

func clinetConncet(addr string, port uint16) (net.Conn, error) {
	return net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
}

func readgo(conn net.Conn) {
	go func() {
		b := make([]byte, 100)
		for {
			n, err := conn.Read(b)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(n, err)
		}
	}()
}

func writego(conn net.Conn) {
	go func() {
		b := make([]byte, 50)
		for {
			n, err := conn.Write(b)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(n, err)
			time.Sleep(time.Second * 5)
		}
	}()
}
