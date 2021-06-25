package locker

import "sync/atomic"

type CASLock struct {
	l uint32
}

func (cl *CASLock) Lock() {
	for {
		if atomic.CompareAndSwapUint32(&cl.l, 0, 1) {
			return
		}
	}
}

func (cl *CASLock) Unlock() {
	for{
		if cl.l == 0 {
			return
		}

		if atomic.CompareAndSwapUint32(&cl.l, 1, 0){
			return
		}
	}
}
