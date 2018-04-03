package timer

import (
	"container/heap"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

const (
	MIN_TIMER_INTERVAL = 1 * time.Millisecond
)

var (
	nextAddSeq uint = 1
)

type Timer struct {
	fireTime time.Time
	interval time.Duration
	callback interface{}
	params	 []interface{}
	repeat   bool
	addseq   uint
}

func (t *Timer) Cancel() {
	t.callback = nil
}

func (t *Timer) IsActive() bool {
	return t.callback != nil
}

type _TimerHeap struct {
	timers []*Timer
}

func (h *_TimerHeap) Len() int {
	return len(h.timers)
}

func (h *_TimerHeap) Less(i, j int) bool {
	//log.Println(h.timers[i].fireTime, h.timers[j].fireTime)
	t1, t2 := h.timers[i].fireTime, h.timers[j].fireTime
	if t1.Before(t2) {
		return true
	}

	if t1.After(t2) {
		return false
	}
	// t1 == t2, making sure Timer with same deadline is fired according to their add order
	return h.timers[i].addseq < h.timers[j].addseq
}

func (h *_TimerHeap) Swap(i, j int) {
	var tmp *Timer
	tmp = h.timers[i]
	h.timers[i] = h.timers[j]
	h.timers[j] = tmp
}

func (h *_TimerHeap) Push(x interface{}) {
	h.timers = append(h.timers, x.(*Timer))
}

func (h *_TimerHeap) Pop() (ret interface{}) {
	l := len(h.timers)
	h.timers, ret = h.timers[:l-1], h.timers[l-1]
	return
}

// Type of callback function
//type CallbackFunc func()

var (
	timerHeap     _TimerHeap
	timerHeapLock sync.Mutex
	isClose 	  bool
)

func init() {
	heap.Init(&timerHeap)
	fmt.Println("start")
	//---
	isClose = false
	StartTicks(MIN_TIMER_INTERVAL)
	
}

// Add a callback which will be called after specified duration
func AddCallback(d time.Duration, callback interface{},params ...interface{}) *Timer {
	
	
	
	t := &Timer{
		fireTime: time.Now().Add(d),
		interval: d,
		callback: callback,
		params : params,
		repeat:   false,
	}
	timerHeapLock.Lock()
	t.addseq = nextAddSeq // set addseq when locked
	nextAddSeq += 1

	heap.Push(&timerHeap, t)
	timerHeapLock.Unlock()
	return t
}

// Add a timer which calls callback periodly
func AddTimer(d time.Duration, callback interface{}) *Timer {
	if d < MIN_TIMER_INTERVAL {
		d = MIN_TIMER_INTERVAL
	}

	t := &Timer{
		fireTime: time.Now().Add(d),
		interval: d,
		callback: callback,
		repeat:   true,
	}
	timerHeapLock.Lock()
	t.addseq = nextAddSeq // set addseq when locked
	nextAddSeq += 1

	heap.Push(&timerHeap, t)
	timerHeapLock.Unlock()
	return t
}

// Tick once for timers
func Tick() {
	now := time.Now()
	timerHeapLock.Lock()

	for {
		if timerHeap.Len() <= 0 {
			break
		}
		
		//fmt.Println( fmt.Sprintln("len:%d",timerHeap.Len()))

		nextFireTime := timerHeap.timers[0].fireTime
		//fmt.Printf(">>> nextFireTime %s, now is %s\n", nextFireTime, now)
		if nextFireTime.After(now) {
			break
		}

		t := heap.Pop(&timerHeap).(*Timer)

		callback := t.callback
		if callback == nil {
			continue
		}

		if !t.repeat {
			t.callback = nil
		}
		// unlock the lock to run callback, because callback may add more callbacks / timers
		timerHeapLock.Unlock()
		runCallback(callback,t.params)
		timerHeapLock.Lock()

		if t.repeat {
			// add Timer back to heap
			t.fireTime = t.fireTime.Add(t.interval)
			if !t.fireTime.After(now) { // might happen when interval is very small
				t.fireTime = now.Add(t.interval)
			}
			t.addseq = nextAddSeq
			nextAddSeq += 1
			heap.Push(&timerHeap, t)
		}
	}
	timerHeapLock.Unlock()
}

// Start the self-ticking routine, which ticks per tickInterval
func StartTicks(tickInterval time.Duration) {
	go selfTickRoutine(tickInterval)
}
func ExitTicks(){
	isClose = true
}

func selfTickRoutine(tickInterval time.Duration) {
	for !isClose{
		time.Sleep(tickInterval)
		Tick()
	}
}

func runCallback(callback interface{},args []interface{}) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Callback %v paniced: %v\n", callback, err)
			debug.PrintStack()
		}
	}()
	
	if len(args) == 3 { 
        callback.(func(interface{},interface{},interface{}))(args[0],args[1],args[2])
    } else if len(args) == 2 { 
        callback.(func(interface{},interface{}))(args[0],args[1])
    }else if len(args) == 1 { 
        callback.(func(interface{}))(args[0])
    } else {
        callback.(func())()
    }
}
