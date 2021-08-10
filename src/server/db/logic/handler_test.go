package logic

import (
	"mmogo/lib/packet"
	"mmogo/network/socket"
	"mmogo/server/db/global"
	"testing"
)

func HandleTest(socket* socket.GameSocket, recvPacket* packet.WorldPacket, ret* packet.WorldPacket)  {

}


func TestHandler(t *testing.T) {
	global.GamePacketHandlerMgr.Reg(16, HandleTest)
	global.GamePacketHandlerMgr.ResetReg(16, HandleTest)
}
