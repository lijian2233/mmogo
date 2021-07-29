package queue

import (
	linklist "github.com/emirpasic/gods/lists/singlylinkedlist"
	"mmogo/lib/locker"
	"runtime"
	"sync"
)

type UnsafeFastQueue struct {
	list [2]*linklist.List
	index int
	cond *sync.Cond
}

func NewUnsafeFastQueue() *UnsafeFastQueue {
	q := &UnsafeFastQueue{}
	q.list[0] = &linklist.List{}
	q.list[1] = &linklist.List{}
	q.index = 0

	if runtime.NumCPU() == 1 {
		q.cond = sync.NewCond(&locker.CASLock{})
	} else {
		q.cond = sync.NewCond(&sync.Mutex{})
	}
	return q
}

func (q *UnsafeFastQueue) Add(value interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	list := q.list[q.index]
	bEmpty := list.Empty()
	list.Add(value)
	if bEmpty {
		q.cond.Signal()
	}
}

func (q *UnsafeFastQueue) Front() interface{} {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	list := q.list[q.index]
	if list.Empty() {
		return nil
	}
	v, _ := list.Get(0)
	return v
}

func (q *UnsafeFastQueue) Pop() interface{} {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	list := q.list[q.index]

	if list.Empty() {
		q.cond.Wait()
	}

	v, _ := list.Get(0)
	return v
}

func (q *UnsafeFastQueue) PopALl() *linklist.List {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	list := q.list[q.index]
	if list.Empty() {
		q.cond.Wait()
	}

	l := q.list[q.index]
	q.index = (q.index + 1) % 2
	return l
}

func (q *UnsafeFastQueue) Signal() {
	q.cond.Signal()
}

