package timer

import (
	"github.com/emirpasic/gods/sets/treeset"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"mmogo/lib"
	"sync"
	"sync/atomic"
	"time"
)

type asyncTimer struct {
	casLock            CASLock
	cond               *sync.Cond
	nextTimerId        uint64
	state              uint32
	pendingTree        *rbt.Tree
	runningIdMapTree   *rbt.Tree
	runningTimeoutTree *rbt.Tree
	cancelQueue        *treeset.Set
	exitCh             chan bool
	once               sync.Once
}

func (timer *asyncTimer) GetMode() int {
	return Mode_Async
}

func (timer *asyncTimer) Init() {
	timer.state = State_Init
	timer.pendingTree = rbt.NewWith(withUint64Compare)
	timer.runningIdMapTree = rbt.NewWith(withUint64Compare)
	timer.cancelQueue = treeset.NewWith(withUint64Compare)
	timer.runningTimeoutTree = rbt.NewWith(timeoutComapre)
	timer.cond = sync.NewCond(&timer.casLock)
	timer.exitCh = make(chan bool, 1)
}

func (timer *asyncTimer) Tick() {

}

func (timer *asyncTimer) run() {
	for {
		if timer.state == State_Exiting {
			timer.state = State_Exited
			timer.clear()
			timer.exitCh <- true
			return
		}
		timer.casLock.Lock()
		if !timer.cancelQueue.Empty() {
			it := timer.cancelQueue.Iterator()
			for it.Next() {
				timer.runningIdMapTree.Remove(it.Value())
				timer.pendingTree.Remove(it.Value())
			}
		}

		if timer.runningIdMapTree.Empty() && timer.pendingTree.Empty() {
			timer.cond.Wait()
		}

		for {
			if !timer.pendingTree.Empty() {
				it := timer.pendingTree.Iterator()
				it.First()
				timer.runningIdMapTree.Put(it.Key(), it.Value())
				timer.runningTimeoutTree.Put(it.Value(), it.Key())
				timer.pendingTree.Remove(it.Key())
			} else {
				break
			}
		}

		timer.casLock.Unlock()

		now := time.Now().UnixNano() / 1000
		it := timer.runningTimeoutTree.Iterator()

		for {
			if it.First() {
				tInfo, _ := it.Key().(*timerInfo)
				if tInfo.timeout > now {
					return
				}
				timer.runningIdMapTree.Remove(tInfo.timerId)
				timer.runningTimeoutTree.Remove(tInfo)
				tInfo.fn()
			} else {
				break
			}
		}

		time.Sleep(time.Microsecond * 10)
	}
}

func (timer *asyncTimer) Start() error {
	err := error(nil)
	timer.once.Do(
		func() {
			if !atomic.CompareAndSwapUint32(&timer.state, State_Init, State_Running) {
				if timer.state == State_Running {
					err = Err_Start_Timer_Is_Running
				}
				err = Err_Start_Timer_Is_Exiting
				return
			}

			go timer.run()
		})
	return err
}

func (timer *asyncTimer) clear() {
	timer.runningIdMapTree = nil
	timer.runningTimeoutTree = nil
	timer.pendingTree = nil
	timer.cancelQueue = nil
	timer.cond = nil
}

func (timer *asyncTimer) Stop() <-chan bool {
	state := atomic.LoadUint32(&timer.state)
	if state == State_Init {
		if atomic.CompareAndSwapUint32(&timer.state, State_Init, State_Exited) {
			timer.clear()
			timer.exitCh <- true
			return timer.exitCh
		}
	}

	if (state == State_Exited) {
		timer.exitCh <- true
		return timer.exitCh
	}

	if state == State_Running {
		if atomic.CompareAndSwapUint32(&timer.state, State_Running, State_Exiting) {
			cond := timer.cond
			if cond != nil {
				cond.Signal()
			}
		}
	}

	return timer.exitCh
}

func (timer *asyncTimer) AddTimer(duration time.Duration, fn func()) uint64 {
	timer.casLock.Lock()
	defer timer.casLock.Unlock()
	timerId := atomic.AddUint64(&timer.nextTimerId, 1)

	tInfo := new(timerInfo)
	tInfo.fn = fn
	tInfo.timerId = timerId
	tInfo.timeout = duration.Microseconds() + time.Now().UnixNano()/1000

	timer.pendingTree.Put(timerId, tInfo)

	return timer.nextTimerId
}

func (timer *asyncTimer) RemoverTimer(timerId uint64) bool {
	timer.casLock.Lock()
	defer timer.casLock.Unlock()
	_, ok := timer.pendingTree.Get(timerId)
	if ok {
		timer.pendingTree.Remove(timerId)
		return true
	}

	//cancel 队列
	timer.cancelQueue.Add(timerId)
	return true
}

func (timer *asyncTimer) GetState() uint32 {
	return timer.state
}
