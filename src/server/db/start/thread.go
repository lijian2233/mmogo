package start

import (
	"fmt"
	"mmogo/lib/packet"
	"mmogo/lib/queue"
	"mmogo/network/socket"
	"mmogo/server/db/global"
	"sync"
)

var WorkThreads = &workThreads{}

type workThreads struct {
	queues []*queue.FastQueue
	wg     sync.WaitGroup
	bStop  bool
	size   uint8
	once   sync.Once
}

func (h *workThreads) Stop() {
	h.bStop = true
	h.wg.Wait()
}

func (h* workThreads) PostPacket(gameSocket *socket.GameSocket, accountId uint32, packet *packet.UtilPacket)  {
	h.queues[int(accountId % uint32(h.size))].Add(&global.GameQueueData{
		Socket:gameSocket,
		Packet:packet,
	})
}

func (h *workThreads) Start() {
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

func (h *workThreads)thread(index int) {
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
			data, _ := d.(*global.GameQueueData)
			fmt.Println(fmt.Sprintf("thread :%d recv packet :%+v", index, data.Packet))
		}

		if h.bStop {
			h.wg.Done()
			return
		}
	}
}
