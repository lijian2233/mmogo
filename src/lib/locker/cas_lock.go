package locker

import (
	"runtime"
	"sync/atomic"
)

var single_cpu bool = false

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
	for {
		if cl.l == 0 {
			return
		}

		if atomic.CompareAndSwapUint32(&cl.l, 1, 0) {
			return
		}
	}
}

func init() {
	if runtime.NumCPU() == 1 {
		single_cpu = true
	}
}
