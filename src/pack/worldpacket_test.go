package pack

import (
	"testing"
)

func TestNewWorldPacket(t *testing.T) {
	packet := NewWorldPacket(16, 100)
	packet.WriteNum(uint32(10000))
	packet.WriteNum(int16(8))
	packet.WriteString("lijian")
}
