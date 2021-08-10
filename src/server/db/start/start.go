package start

import (
	_interface "mmogo/interface"
	"mmogo/lib/packet"
	"mmogo/network/socket"
	"mmogo/server/db/global"
	"net"
)

func HandleAcceptConn(conn net.Conn) {
	socket.NewConnSocket(conn,
		socket.WithGameSendBuffSize(512*1024),
		socket.WithGameRecivBuffSize(10*1024),
		socket.WithGameLog(global.Log),
		socket.WithGameHandlePacket(func(binaryPacket _interface.BinaryPacket) {
			p, ok := binaryPacket.(*packet.WorldPacket)
			if ok {
				HandleThreads.PostPacket(p.GetSeqNo(), p)
			}
		}),
	)
}




