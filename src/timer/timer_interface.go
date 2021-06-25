package timer

import (
	"errors"
	"time"
)

const (
	State_Init    = 1
	State_Running = 2
	State_Exiting = 3
	State_Exited  = 4
	Mode_Sync     = 1
	Mode_Async    = 2
)

type timerInfo struct {
	fn      func()
	timerId uint64
	timeout int64
}

var Err_Start_Timer_Is_Running = errors.New("timer is running")
var Err_Start_Timer_Is_Exiting = errors.New("timer is exited or exiting")

type Timer interface {
	AddTimer(duration time.Duration, fn func()) uint64
	RemoverTimer(timerId uint64) bool
	Start() error
	Stop() <-chan bool
	GetState() uint32
	Init()
	Tick()
	GetMode() int
}

func withUint64Compare(a, b interface{}) int {
	a1, _ := a.(uint64)
	b1, _ := b.(uint64)

	if a1 < b1 {
		return -1
	}

	if a1 == b1 {
		return 0
	}

	return 1
}

func NewSyncTimer() Timer {
	timer := new(syncTimer)
	timer.Init()
	return timer
}

func NewAsyncTimer() Timer {
	timer := new(asyncTimer)
	timer.Init()
	return timer
}

// Should return a number:
//    negative , if a < b
//    zero     , if a == b
//    positive , if a > b

func timeoutComapre(a, b interface{}) int {
	t1, _ := a.(*timerInfo)
	t2, _ := b.(*timerInfo)

	if t1.timeout < t2.timeout {
		return -1
	}

	if t1.timeout > t2.timeout {
		return 1
	}

	if t1.timerId < t2.timerId {
		return -1
	}

	if t1.timerId < t2.timerId {
		return 1
	}
	return 0
}
