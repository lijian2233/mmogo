package locker

import "sync"

type MutexLock struct {
	m sync.Mutex
}

func (ml *MutexLock) Lock() {
	ml.m.Lock()
}

func (ml *MutexLock) Unlock() {
	ml.m.Unlock()
}
