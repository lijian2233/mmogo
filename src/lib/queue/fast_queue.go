package queue

import (
	linklist "github.com/emirpasic/gods/lists/singlylinkedlist"
	"mmogo/lib/locker"
	"runtime"
	"sync"
)

type FastQueue struct {
	list *linklist.List
	cond *sync.Cond
}

func NewFastQueue() *FastQueue {
	q := &FastQueue{}
	q.list = &linklist.List{}

	if runtime.NumCPU() == 1 {
		q.cond = sync.NewCond(&locker.CASLock{})
	} else {
		q.cond = sync.NewCond(&sync.Mutex{})
	}
	return q
}

func (q *FastQueue) Add(value interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	bEmpty := q.list.Empty()
	q.list.Add(value)
	if bEmpty {
		q.cond.Signal()
	}
}

func (q *FastQueue) Front() interface{} {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.list.Empty() {
		return nil
	}
	v, _ := q.list.Get(0)
	return v
}

func (q *FastQueue) Pop() interface{} {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.list.Empty() {
		q.cond.Wait()
	}

	if q.list.Empty() {
		return nil
	}
	v, _ := q.list.Get(0)
	return v
}

func (q *FastQueue) PopALl() *linklist.List {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.list.Iterator()

	if q.list.Empty() {
		q.cond.Wait()
	}

	if q.list.Empty() {
		return nil
	}

	l := q.list
	q.list = &linklist.List{}
	return l
}

/*
* if queue data empty, then will return nil
 */
func (q *FastQueue) All() *linklist.List {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.list.Empty() {
		return nil
	}

	l := q.list
	q.list = &linklist.List{}
	return l
}

func (q *FastQueue) Signal() {
	q.cond.Signal()
}
