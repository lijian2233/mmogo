package pack

import (
	"fmt"
	"testing"
)

func TestNewWorldPacket(t *testing.T) {
	packet := NewWorldPacket(16, 100)
	packet.WriteNum(uint32(10000))
	packet.WriteNum(int16(8))
	packet.WriteString("lijian")

	fmt.Println(packet.ReadUint32(),packet.ReadUint16(),packet.readString())
}
