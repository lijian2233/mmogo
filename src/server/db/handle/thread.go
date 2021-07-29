package handle

import (
	"fmt"
	_interface "mmogo/interface"
	"mmogo/lib/packet"
	"mmogo/lib/queue"
	"mmogo/server/db/global"
	"sync"
)

var HandleThreads *handleThreads

type handleThreads struct {
	queues []*queue.FastQueue
	wg     sync.WaitGroup
	bStop  bool
	size   uint8
}

func (h *handleThreads) Stop() {
	h.bStop = true
	h.wg.Wait()
}

func (h* handleThreads) PostPacket(accountId uint64, packet *packet.WorldPacket)  {
	h.queues[int(accountId % uint64(h.size))].Add(packet)
}

func (h *handleThreads) Start() {
	threadNum := global.Conf.HandlerThreadNum
	if threadNum == 0 {
		threadNum = 4
	}

	if threadNum > 128 {
		threadNum = 128
	}

	HandleThreads = &handleThreads{}
	HandleThreads.queues = make([]*queue.FastQueue, threadNum, threadNum)

	HandleThreads.wg.Add(int(threadNum))
	h.size = threadNum
	for i := 0; i < int(threadNum); i++ {
		HandleThreads.queues[i] = queue.NewFastQueue()
		go thread(i)
	}
}

func thread(index int) {
	queue := HandleThreads.queues[index]

	for {
		if HandleThreads.bStop {
			HandleThreads.wg.Done()
			return
		}

		list := queue.PopALl()
		for{
			if list.Empty() {
				break
			}

			d, _ := list.Get(0)
			packet, _ := d.(_interface.BinaryPacket)
			fmt.Println(packet)
		}

		if HandleThreads.bStop {
			HandleThreads.wg.Done()
			return
		}
	}
}
