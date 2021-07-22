package timer

import (
	"fmt"
	"testing"
	"time"
)

func test1(a, b int) int {
	return a + b
}

func test2(a, b int)  int{
	return a - b
}

func TestTimer_SyncTimer(t *testing.T) {
	timer := NewSyncTimer()
	timer.Start()

	timer.AddTimer(time.Microsecond, func() {
		test1(10, 15)
	})

	timer.AddTimer(time.Microsecond, func() {
		test1(10, 15)
	})


	time.Sleep(time.Microsecond*10)

	timer.Tick()

	f := <-timer.Stop()
	fmt.Println(f)
}

func TestTimer_AsyncTimer(t *testing.T)  {
	timer := NewAsyncTimer()
	timer.Start()

	timer.AddTimer(time.Microsecond, func() {
		test1(10, 15)
	})

	time.Sleep(time.Microsecond*10)

	f :=<-timer.Stop()
	fmt.Println(f)

}
