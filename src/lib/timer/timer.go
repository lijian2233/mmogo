package timer

/*
import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"github.com/pkg/errors"
	"mmogo/common/locker"
	"sync"
	"sync/atomic"
	"time"
)

func test() {
	rbt.NewWith(utils.UInt64Comparator)
}

var Err_Start_Timer_Is_Running = errors.New("timer is running")
var Err_Start_Timer_Is_Exiting = errors.New("timer is exited or exiting")
//1:初始化 1:运行中 2:退出中 3:已退出
const (
	State_Init    = 1
	State_Running = 2
	State_Exiting = 3
	State_Exited  = 4
	Mode_Sync     = 1
	Mode_Asyn     = 2
)

type callbackFn struct {
	class interface{}
	fn    interface{}
	param []interface{}
}

type Timer struct {
	nextId   uint64
	mutex    sync.Mutex
	state    uint32
	mode     uint32
	treeLock locker.CASLock //CAS lock
	cond*    sync.Cond
	treeMap  *rbt.Tree
	exitCh   chan bool
}

func NewTimer() *Timer {
	timer := Timer{state: State_Init}
	timer.treeMap = rbt.NewWithIntComparator()
	timer.treeLock = 0
	timer.nextId = 0
	timer.mode = Mode_Sync
	return &timer
}

func (timer* Timer) Tick()  {

}


func (timer *Timer) run() {
	for{
		timer.treeLock.Lock()
		if timer.treeMap.Size() == 0{
			timer.cond.Wait()
		}
	}
	timer.mutex.Lock()
	timer.state = State_Exited
	timer.mutex.Unlock()

	timer.exitCh <- true
}

func (timer *Timer) canStart() error {
	if timer.state != State_Init {
		if timer.state == State_Running {
			return Err_Start_Timer_Is_Running
		} else {
			return Err_Start_Timer_Is_Exiting
		}
	}
	return nil
}

func (timer *Timer) Start(mode uint32) error {
	err := timer.canStart()
	if err != nil {
		return err
	}

	if !atomic.CompareAndSwapUint32(&timer.state, State_Init, State_Running) {
		err := timer.canStart()
		if err != nil {
			return err
		}
		//多协程会不会有状态不一致，待测试
		return Err_Start_Timer_Is_Exiting
	}

	timer.mode = mode
	if mode == Mode_Asyn {
		timer.cond = sync.NewCond(&timer.treeLock)
		go timer.run()
	}
	return nil
}

func (timer* Timer) syncStop() {
}

func (timer* Timer) asyncStop()  {

}

func (timer *Timer) Stop() {
	if timer.mode != Mode_Sync {
		timer.mutex.Lock()
	}

	if timer.state == State_Init {
		timer.state = State_Exiting
		if (timer.mode != Mode_Sync) {
			timer.mutex.Unlock()
		}
		timer.exitCh <- true
		return
	}
	{
		timer.mutex.Unlock()
		if timer.state == State_Exited {
			timer.state = State_Exited
			return
		}

		if timer.state == State_Running {
			timer.state = State_Exiting
		}
	}
}

func (timer *Timer) State() uint32 {
	return timer.state
}

func (timer *Timer) genTimerId() uint64 {
	return atomic.AddUint64(&timer.nextId, 1)
}

*/

//不校验interface值为空, 只校验interface 类型为nil
/*只判断类型1
* var a interface{} = nil // tab = nil, data = nil
* var b interface{} = (*int)(nil) // tab 包含 *int 类型信息, data = nil
 */

/*
func (timer *Timer) addTimer(duration time.Duration, class interface{},
	fn interface{}, params ...interface{}) {

	cfn := new(callbackFn)
	cfn.class = class
	cfn.fn = fn
	cfn.param = params

	timerId := timer.genTimerId()

	if timer.mode == Mode_Sync {
		timer.treeMap.Put(timerId, cfn)
	} else {
		timer.treeLock.Lock()
		defer timer.treeLock.Unlock()
		timer.treeMap.Put(timerId, cfn)
		if timer.treeMap.Size() == 1{
			timer.cond.Signal()
		}
	}
}

func (timer *Timer) cancelTimer(timerId uint64) {
	if timer.mode == Mode_Sync {
		timer.treeMap.Remove(timerId)
	} else {
		timer.treeLock.Lock()
		defer timer.treeLock.Unlock()
		timer.treeMap.Remove(timerId)
	}
}
*/
