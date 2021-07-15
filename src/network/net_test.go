package network

import (
	"fmt"
	"mmogo/common/logger"
	"mmogo/pack"
	"net"
	"testing"
	"time"
)

var log *logger.Logger
func init()  {
	log, _ = logger.NewLogger(logger.DefaultConfig())
}

func parsePacket(packet *pack.WorldPacket)  {
	switch packet.GetOpcode() {
	case 201:{
		a, b := packet.ReadUint32(),packet.ReadUint32()
		if a != 1 || b !=2 {
			panic("201 error")
		}
		break
	}
	case 202:
		{
			a, s, b := packet.ReadUint32(), packet.ReadString(), packet.ReadUint32()
			if a != 3 || b != 4 {
				panic("202 error")
			}
			fmt.Println("202 packet ", s)
			break
		}
	case 203:
		{
			a, s, b := packet.ReadUint32(), packet.ReadBytes(), packet.ReadUint32()
			if a != 5 || b != 6 {
				panic("203 error")
			}
			fmt.Println(s)
			break
		}
	}
}



func handleAccept(conn net.Conn) {
	 NewConnSocket(conn, 512, 256, func(packet *pack.WorldPacket) {
		fmt.Println(packet.GetOpcode(), packet.GetSize())
	})
}

func TestNew(t *testing.T)  {
	go TestNewListenSocket(t)
	time.Sleep(time.Second*5)
	go TestClient(t)
	time.Sleep(time.Minute*10)
}

func TestNewListenSocket(t *testing.T) {

	l, err := NewListenSocket("127.0.0.1", 8080, *log, handleAccept)
	if err != nil {
		fmt.Println(err)
		return
	}

	go l.Start()
}

func TestClient(t *testing.T)  {
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", 8080))
	if err != nil {
		fmt.Println("TestClient err ", err)
		return
	}

	cli, err := NewConnSocket(c, 10*1024, 10*1024, func(packet *pack.WorldPacket) {
		
	})

	if err != nil{
		fmt.Println("TestClient err ", err)
	}

	for i:=0; i<100; i++{
		p1 := pack.NewWorldPacket(201, 100)
		p2 := pack.NewWorldPacket(202, 100)
		p3 := pack.NewWorldPacket(203, 100)

		p1.WriteNum(uint32(1))
		p2.WriteNum(uint32(2))

		p2.WriteNum(uint32(3))
		p2.WriteString("lijian")
		p2.WriteNum(uint32(4))

		p3.WriteNum(uint32(3))
		p3.WriteString("lijian")
		p3.WriteNum(uint32(4))

		cli.SendMsg(p1.GetContent())
		cli.SendMsg(p2.GetContent())
		cli.SendMsg(p3.GetContent())
	}
}