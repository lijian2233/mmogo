package network

import (
	"fmt"
	"mmogo/common/logger"
	"mmogo/pack"
	"net"
	"sync"
	"testing"
	"time"
)

var log *logger.Logger

func init() {
	log, _ = logger.NewLogger(logger.DefaultConfig())
}

var tt *testing.T

func parsePacket(packet *pack.WorldPacket) {
	switch packet.GetOpcode() {
	case 201:
		{
			a, b := packet.ReadUint32(), packet.ReadUint32()
			if a != 1 || b != 2 {
				panic("201 error")
			}
			log.Info("201", a, b)
			break
		}
	case 202:
		{
			a, s, b := packet.ReadUint32(), packet.ReadString(), packet.ReadUint32()
			if a != 3 || b != 4 {
				panic("202 error")
			}
			log.Info("202", a, s, b)
			break
		}
	case 203:
		{
			a, s, b := packet.ReadUint32(), packet.ReadString(), packet.ReadUint32()
			if a != 5 || b != 6 {
				panic("203 error")
			}
			log.Info("203", a, s, b)
			break
		}
	}
}

func handleAccept(conn net.Conn) {
	 NewConnSocket(conn, 512, 256, func(packet *pack.WorldPacket) {
		fmt.Println(packet.GetOpcode(), packet.GetSize())
		parsePacket(packet)
	}, log)

}

func TestNew(t *testing.T) {
	tt = t
	go TestNewListenSocket(t)
	time.Sleep(time.Second * 5)
	go TestClient(t)
	time.Sleep(time.Minute * 10)
}

func TestNewListenSocket(t *testing.T) {
	l, err := NewListenSocket("127.0.0.1", 8080, *log, handleAccept)
	if err != nil {
		fmt.Println(err)
		return
	}

	go l.Start()
}

type CondTest struct {
	lock sync.Mutex
	cond *sync.Cond
}

var condTest *CondTest

func Init() {
	condTest = &CondTest{}
	condTest.cond = sync.NewCond(&condTest.lock)
}

func g1() {
	condTest.lock.Lock()
	time.Sleep(time.Second * 3)
	condTest.cond.Broadcast()
	condTest.lock.Unlock()

	condTest.lock.Lock()
	condTest.lock.Unlock()
}

func g2() {
	condTest.lock.Lock()
	condTest.cond.Wait()
	condTest.lock.Unlock()
}

func TestCond(t *testing.T) {
	Init()
	go g2()
	time.Sleep(time.Second * 1)
	go g1()

	time.Sleep(time.Second * 20)
}

func TestClient(t *testing.T) {
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", 8080))
	if err != nil {
		fmt.Println("TestClient err ", err)
		return
	}

	cli, err := NewConnSocket(c, 10*1024, 10*1024, func(packet *pack.WorldPacket) {

	}, log)

	if err != nil {
		fmt.Println("TestClient err ", err)
	}

	for i := 0; i < 100; i++ {
		p1 := pack.NewWorldPacket(201, 100)
		p1.WriteNum(uint32(1))
		p1.WriteNum(uint32(2))
		cli.SendPacket(p1)

		p2 := pack.NewWorldPacket(202, 100)
		p2.WriteNum(uint32(3))
		p2.WriteString("lijian")
		p2.WriteNum(uint32(4))
		cli.SendPacket(p2)

		p3 := pack.NewWorldPacket(203, 100)
		p3.WriteNum(uint32(5))
		p3.WriteString("lijian")
		p3.WriteNum(uint32(6))
		cli.SendPacket(p3)
	}
}
