package timer

import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"sync"
	"sync/atomic"
	"time"
)

type syncTimer struct {
	nextTimerId uint64
	state       uint32
	idMap       *rbt.Tree
	timeoutMap  *rbt.Tree
	exitCh      chan bool
	once        sync.Once
}

func (timer *syncTimer) Init() {
	timer.state = State_Init
	timer.idMap = rbt.NewWith(withUint64Compare)
	timer.timeoutMap = rbt.NewWith(timeoutComapre)
	timer.exitCh = make(chan bool, 1)
}

func (timer *syncTimer) GetMode() int {
	return Mode_Sync
}

func (timer *syncTimer) Start() error {
	timer.once.Do(
		func() {
			if !atomic.CompareAndSwapUint32(&timer.state, State_Init, State_Running) {
				panic("sync timer running multi go routine")
			}
		})

	return nil
}

func (timer *syncTimer) Stop() <-chan bool {
	if timer.state == State_Init {
		if !atomic.CompareAndSwapUint32(&timer.state, State_Init, State_Exited) {
			panic("sync timer running multi go routine")
		}
	}

	if timer.state == State_Exited {
		timer.exitCh <- true
		return timer.exitCh
	}

	if timer.state == State_Running {
		if !atomic.CompareAndSwapUint32(&timer.state, State_Running, State_Exited) {
			panic("sync timer running multi go routine")
		}
		timer.timeoutMap = nil
		timer.idMap = nil
	}
	timer.exitCh <- true
	return timer.exitCh
}

func (timer *syncTimer) GetState() uint32 {
	return timer.state
}

func (timer *syncTimer) AddTimer(duration time.Duration, fn func()) uint64 {
	timer.nextTimerId = timer.nextTimerId + 1
	timerId := timer.nextTimerId
	tInfo := new(timerInfo)
	tInfo.fn = fn
	tInfo.timerId = timerId
	now := time.Now().UnixNano() / 1000
	tInfo.timeout = now + duration.Microseconds()

	timer.idMap.Put(timerId, tInfo)
	timer.timeoutMap.Put(tInfo, tInfo.timerId)
	return timerId
}

func (timer *syncTimer) RemoverTimer(timerId uint64) bool {
	t, ok := timer.idMap.Get(timerId)
	if ok {
		timer.idMap.Remove(timerId)
		timer.timeoutMap.Remove(t)
	}
	return true
}

func (timer *syncTimer) Tick() {
	now := time.Now().UnixNano() / 1000
	for {
		if !timer.timeoutMap.Empty() {
			it := timer.timeoutMap.Iterator()
			it.First()
			tInfo, _ := it.Key().(*timerInfo)
			if tInfo.timeout > now {
				return
			}
			timer.RemoverTimer(tInfo.timerId)
			tInfo.fn()
		}else{
			break
		}
	}
}

