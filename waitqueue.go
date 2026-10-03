package nsync

import (
	"context"
	"sync"
	"time"
)

// waitQueue is a FIFO list of parked waiters guarded by mu.
type waitQueue struct {
	mu sync.Mutex
	waitList
	spare *waiter // a finished waiter kept for reuse; avoids the pool
}

// take returns a waiter whose condition variable uses q.mu. q.mu must be held.
func (q *waitQueue) take() *waiter {
	if w := q.spare; w != nil {
		q.spare = nil
		return w
	}
	return getWaiter(&q.mu)
}

// keep recycles a waiter from take that is no longer queued and whose
// callback, if any, has finished. q.mu must be held.
func (q *waitQueue) keep(w *waiter) {
	if q.spare != nil {
		putWaiter(w)
		return
	}
	w.reset()
	q.spare = w
}

// waitList is a doubly linked list of waiters; its lock is the waiters' cond.L.
type waitList struct {
	head *waiter
	tail *waiter
	n    int
}

// A waiter parks on its own condition variable, so a release, completion, or
// cancellation wakes exactly the goroutine it concerns. Waiters are pooled;
// fields other than fireFunc are protected by cond.L.
type waiter struct {
	cond     sync.Cond
	list     *waitList // list holding the waiter while queued
	prev     *waiter
	next     *waiter
	since    time.Time // when the waiter queued, for starvation detection
	fireFunc func()    // cached method value of fire; avoids an allocation per timer
	queued   bool
	woken    bool // asked to retry; keeps its position in the queue
	granted  bool // a permit or a completion was handed over directly
	canceled bool // removed by its timer or context before being granted
	fired    bool // its timer or context callback has finished
}

var waiterPool = sync.Pool{New: func() any {
	w := new(waiter)
	w.fireFunc = w.fire
	return w
}}

func getWaiter(mu *sync.Mutex) *waiter {
	w := waiterPool.Get().(*waiter)
	w.cond.L = mu
	return w
}

// putWaiter recycles a waiter that is no longer queued and whose callback, if
// any, has finished. Signals to it have all been consumed, so its condition
// variable has no pending tickets.
func putWaiter(w *waiter) {
	w.cond.L = nil
	w.reset()
	waiterPool.Put(w)
}

func (w *waiter) reset() {
	w.list = nil
	w.since = time.Time{}
	w.woken, w.granted, w.canceled, w.fired = false, false, false, false
}

func (l *waitList) pushBack(w *waiter) {
	w.list = l
	w.prev, w.next = l.tail, nil
	if l.tail == nil {
		l.head = w
	} else {
		l.tail.next = w
	}
	l.tail = w
	l.n++
	w.queued = true
}

func (l *waitList) remove(w *waiter) {
	if w.prev == nil {
		l.head = w.next
	} else {
		w.prev.next = w.next
	}
	if w.next == nil {
		l.tail = w.prev
	} else {
		w.next.prev = w.prev
	}
	w.prev, w.next = nil, nil
	l.n--
	w.queued = false
}

// grantAll hands a completion to every waiter in the list.
func (l *waitList) grantAll() {
	for w := l.head; w != nil; w = l.head {
		l.remove(w)
		w.granted = true
		w.cond.Signal()
	}
}

// wakeOne asks the first waiter that is not already retrying to retry.
func (l *waitList) wakeOne() {
	for w := l.head; w != nil; w = w.next {
		if !w.woken {
			w.woken = true
			w.cond.Signal()
			return
		}
	}
}

// fire runs in its own goroutine when the waiter's timer expires or its
// context is done. A waiter that was already granted keeps its grant.
func (w *waiter) fire() {
	mu := w.cond.L
	mu.Lock()
	if w.queued && !w.granted {
		w.list.remove(w)
		w.canceled = true
	}
	w.fired = true
	w.cond.Signal()
	mu.Unlock()
}

// waitCancel is the cancellation source armed for one wait, if any.
type waitCancel struct {
	timer *time.Timer
	stop  func() bool
}

// arm starts the waiter's timer, or registers it with ctx. Timers are never
// reused: each one belongs to the testing/synctest bubble that created it.
func (w *waiter) arm(timeout time.Duration, ctx context.Context) (c waitCancel) {
	if timeout > 0 {
		c.timer = time.AfterFunc(timeout, w.fireFunc)
	} else if ctx != nil && ctx.Done() != nil {
		c.stop = context.AfterFunc(ctx, w.fireFunc)
	}
	return c
}

// disarm requires cond.L to be held and returns with it held. If the callback
// has already started, it waits for the callback to finish with the waiter.
func (w *waiter) disarm(c waitCancel) {
	switch {
	case c.timer != nil:
		if c.timer.Stop() {
			return
		}
	case c.stop != nil:
		if c.stop() {
			return
		}
	default:
		return
	}
	for !w.fired {
		w.cond.Wait()
	}
}
