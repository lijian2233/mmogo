package global

import (
	"fmt"
	"mmogo/lib/packet"
	"mmogo/network/socket"
	"runtime"
	"sync"
)

type GameQueueData struct {
	Socket *socket.GameSocket
	Packet *packet.WorldPacket
}

type GamePacketHandler func(socket *socket.GameSocket, recvPacket *packet.WorldPacket, ret *packet.WorldPacket)

var GamePacketHandlerMgr = &gamePacketHandlerMgr{
	id2Handler:  map[uint16]GamePacketHandler{},
	id2RegStack: map[uint16]string{},
}

type gamePacketHandlerMgr struct {
	mutex       sync.Mutex
	id2Handler  map[uint16]GamePacketHandler
	id2RegStack map[uint16]string
}

func (m *gamePacketHandlerMgr) Reg(op uint16, handler GamePacketHandler)  {
	_, file, line, _ := runtime.Caller(0)
	rInfo := fmt.Sprintf("file:%s line:%d", file, line)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.id2Handler[op]; ok {
		regInfo, _ := m.id2RegStack[op]
		pInfo := fmt.Sprintf("op :%d has reg in :[%s]", op, regInfo)
		panic(pInfo)

	}else{
		m.id2Handler[op] = handler
		m.id2RegStack[op] = rInfo
	}
}

func (m *gamePacketHandlerMgr) ResetReg(op uint16, handler GamePacketHandler)  {
	_, file, line, _ := runtime.Caller(0)
	rInfo := fmt.Sprintf("file:%s line:%d", file, line)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.id2Handler[op]; ok {
		if Log != nil {
			Log.Infof("replace reg:%d", op)
		}
	}

	m.id2Handler[op] = handler
	m.id2RegStack[op] = rInfo
}