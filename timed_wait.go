package nsync

import (
	"sync"
	"time"
)

// timedWait reuses callback state, but never a timer: timers may belong to
// different testing/synctest bubbles. Its fields are protected by cond.L.
type timedWait struct {
	cond     *sync.Cond
	expired  bool
	callback func()
}

var timedWaitPool = sync.Pool{New: func() any {
	w := new(timedWait)
	w.callback = w.expire
	return w
}}

// startTimedWait requires cond.L to be held.
func startTimedWait(cond *sync.Cond, deadline time.Time) (*timedWait, *time.Timer) {
	w := timedWaitPool.Get().(*timedWait)
	timeout := time.Until(deadline)
	if timeout <= 0 {
		timedWaitPool.Put(w)
		return nil, nil
	}
	w.cond = cond
	w.expired = false
	return w, time.AfterFunc(timeout, w.callback)
}

func (w *timedWait) expire() {
	c := w.cond
	c.L.Lock()
	w.expired = true
	c.Broadcast()
	// Do not touch w after unlocking: the waiter may now recycle it.
	c.L.Unlock()
}

// stop requires cond.L to be held and returns with it held. If the callback
// already started, wait for its critical section before recycling its state.
func (w *timedWait) stop(timer *time.Timer) {
	if !timer.Stop() {
		for !w.expired {
			w.cond.Wait()
		}
	}
	w.cond = nil
	timedWaitPool.Put(w)
}
