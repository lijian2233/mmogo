package start

import (
	"fmt"
	_interface "mmogo/interface"
	"mmogo/lib/packet"
	"mmogo/lib/queue"
	"mmogo/server/db/global"
	"sync"
)

var HandleThreads = &handleThreads{}

type handleThreads struct {
	queues []*queue.FastQueue
	wg     sync.WaitGroup
	bStop  bool
	size   uint8
	once   sync.Once
}

func (h *handleThreads) Stop() {
	h.bStop = true
	h.wg.Wait()
}

func (h* handleThreads) PostPacket(accountId uint32, packet *packet.WorldPacket)  {
	h.queues[int(accountId % uint32(h.size))].Add(packet)
}

func (h *handleThreads) Start() {
	h.once.Do(func() {
		threadNum := global.Conf.HandlerThreadNum
		if threadNum == 0 {
			threadNum = 4
		}

		if threadNum > 128 {
			threadNum = 128
		}

		h.queues = make([]*queue.FastQueue, threadNum, threadNum)

		h.wg.Add(int(threadNum))
		h.size = threadNum
		for i := 0; i < int(threadNum); i++ {
			h.queues[i] = queue.NewFastQueue()
			go h.thread(i)
		}
	})
}

func (h *handleThreads)thread(index int) {
	queue := h.queues[index]

	for {
		if h.bStop {
			h.wg.Done()
			return
		}

		list := queue.PopALl()
		for{
			if list.Empty() {
				break
			}

			d, _ := list.Get(0)
			list.Remove(0)
			packet, _ := d.(_interface.BinaryPacket)
			fmt.Println(fmt.Sprintf("thread :%d recv packet :%+v", index, packet))
		}

		if h.bStop {
			h.wg.Done()
			return
		}
	}
}
