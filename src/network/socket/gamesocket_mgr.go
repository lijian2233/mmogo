package socket

import (
	"mmogo/lib/locker"
	"net"
)

type GameSocketMgr struct {
	sockets map[*GameSocket]net.Conn
	locker  locker.CASLock
}

func NewGameSocketMgr() *GameSocketMgr {
	mgr := new (GameSocketMgr)
	mgr.sockets = map[*GameSocket]net.Conn{}
	return mgr
}

func (gm *GameSocketMgr) AddSocket(socket *GameSocket) {
	gm.locker.Lock()
	defer gm.locker.Unlock()

	gm.sockets[socket] = socket.conn
}

func (gm *GameSocketMgr) RemoveSocket(socket *GameSocket) {
	gm.locker.Lock()
	defer gm.locker.Unlock()

	delete(gm.sockets, socket)
}

func (gm *GameSocketMgr) ClearGameSocket() {
	gm.locker.Lock()
	defer gm.locker.Unlock()

	gm.sockets = map[*GameSocket]net.Conn{}
}

func (gm* GameSocketMgr) ForEach(fn func(socket *GameSocket))  {
	gm.locker.Lock()
	m := map[*GameSocket]net.Conn{}
	for k, v := range(gm.sockets){
		m[k] = v
	}
	defer gm.locker.Unlock()

	for k, _ := range(m){
		fn(k)
	}
}