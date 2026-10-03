package nsync

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrAborted is returned by DoContext when the group has been aborted.
var ErrAborted = errors.New("nsync: control wait group aborted")

// ControlWaitGroup runs tasks with a limit on concurrent goroutines.
// Use NewControlWaitGroup to initialize it. A ControlWaitGroup must not be
// copied after first use.
type ControlWaitGroup struct {
	// state holds the number of running tasks and flags. Admission and
	// completion use it alone unless submitters or Wait calls are parked.
	state atomic.Uintptr
	limit uintptr // immutable after construction
	// queue holds Do calls waiting for a slot, in arrival order.
	queue   waitQueue
	waiters waitList // Wait calls; guarded by queue.mu
}

const (
	cwgQueued  uintptr = 1 << iota // submitters are queued: freed slots are handed to them
	cwgWaiting                     // Wait calls are parked
	cwgAborted
)

const (
	cwgShift         = 3
	cwgUnit  uintptr = 1 << cwgShift // one running task
)

// NewControlWaitGroup creates a group. It panics unless poolSize is positive.
func NewControlWaitGroup(poolSize int) *ControlWaitGroup {
	if poolSize <= 0 {
		panic("nsync: worker limit must be positive")
	}
	return &ControlWaitGroup{limit: min(uintptr(poolSize), ^uintptr(0)>>cwgShift)}
}

// Do waits for a free slot and starts userFunc in a new goroutine, returning true.
// If the group is aborted before admission, Do returns false without running
// userFunc. Submission to an empty group must precede Wait.
func (cwg *ControlWaitGroup) Do(userFunc func()) bool {
	return cwg.state.Load()&cwgAborted == 0 && cwg.do(userFunc)
}

func (cwg *ControlWaitGroup) do(userFunc func()) bool {
	s := cwg.state.Load()
	if (s&(cwgQueued|cwgAborted) != 0 || s>>cwgShift >= cwg.limit || !cwg.state.CompareAndSwap(s, s+cwgUnit)) &&
		cwg.admitSlow(context.Background()) != nil {
		return false
	}
	go func() {
		defer cwg.finish()
		userFunc()
	}()
	return true
}

// DoContext is like Do, but stops waiting for a slot when ctx is done. It
// returns nil once userFunc has been started, ErrAborted if the group was
// aborted first, or ctx.Err(). If ctx is already done, DoContext does not
// start userFunc.
func (cwg *ControlWaitGroup) DoContext(ctx context.Context, userFunc func()) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s := cwg.state.Load()
	if s&(cwgQueued|cwgAborted) != 0 || s>>cwgShift >= cwg.limit || !cwg.state.CompareAndSwap(s, s+cwgUnit) {
		if err := cwg.admitSlow(ctx); err != nil {
			return err
		}
	}
	go func() {
		defer cwg.finish()
		userFunc()
	}()
	return nil
}

// admitSlow waits for a slot. Submitters are admitted in arrival order.
func (cwg *ControlWaitGroup) admitSlow(ctx context.Context) error {
	cwg.queue.mu.Lock()
	for {
		s := cwg.state.Load()
		if s&cwgAborted != 0 {
			cwg.queue.mu.Unlock()
			return ErrAborted
		}
		if s&cwgQueued == 0 && s>>cwgShift < cwg.limit {
			if cwg.state.CompareAndSwap(s, s+cwgUnit) {
				cwg.queue.mu.Unlock()
				return nil
			}
			continue
		}
		if err := ctx.Err(); err != nil {
			cwg.queue.mu.Unlock()
			return err
		}
		if s&cwgQueued != 0 || cwg.state.CompareAndSwap(s, s|cwgQueued) {
			break
		}
	}
	w := cwg.queue.take()
	cwg.queue.pushBack(w)
	c := w.arm(0, ctx)
	for !w.granted && !w.canceled {
		w.cond.Wait()
	}
	w.disarm(c)
	var err error
	if !w.granted {
		// Removed by Abort or by its context.
		if err = ErrAborted; cwg.state.Load()&cwgAborted == 0 {
			err = ctx.Err()
		}
		cwg.settle()
	}
	cwg.queue.keep(w)
	cwg.queue.mu.Unlock()
	return err
}

func (cwg *ControlWaitGroup) finish() {
	for {
		s := cwg.state.Load()
		if s&(cwgQueued|cwgWaiting) != 0 {
			cwg.finishSlow()
			return
		}
		if cwg.state.CompareAndSwap(s, s-cwgUnit) {
			return
		}
	}
}

func (cwg *ControlWaitGroup) finishSlow() {
	cwg.queue.mu.Lock()
	if h := cwg.queue.head; h != nil && cwg.state.Load()&cwgAborted == 0 {
		// Hand the slot to the longest-waiting submitter. The running count
		// is unchanged, so no newcomer can take the slot in between. The
		// queued flag stays set even if the queue is now empty: the group is
		// still full, the submitter usually queues again at once, and the
		// next completion clears the flag. (While the group is not aborted,
		// the flag always implies a full group; a canceled submitter can
		// leave it set too, until it settles.)
		cwg.queue.remove(h)
		h.granted = true
		h.cond.Signal()
	} else {
		cwg.state.Add(^(cwgUnit - 1))
		cwg.settle()
	}
	cwg.queue.mu.Unlock()
}

// settle clears stale flags and releases Wait calls once no task is running
// or pending. queue.mu must be held.
func (cwg *ControlWaitGroup) settle() {
	for {
		s := cwg.state.Load()
		ns := s
		if cwg.queue.head == nil {
			ns &^= cwgQueued
		}
		idle := ns>>cwgShift == 0 && ns&cwgQueued == 0
		if idle {
			ns &^= cwgWaiting
		}
		if ns == s || cwg.state.CompareAndSwap(s, ns) {
			if idle {
				cwg.waiters.grantAll()
			}
			return
		}
	}
}

// Abort unblocks pending Do calls and permanently rejects new tasks.
// Already admitted tasks may continue running; use Wait to wait for them.
// Repeated calls are safe.
func (cwg *ControlWaitGroup) Abort() {
	for {
		s := cwg.state.Load()
		if s&cwgAborted != 0 || cwg.state.CompareAndSwap(s, s|cwgAborted) {
			break
		}
	}
	cwg.queue.mu.Lock()
	for w := cwg.queue.head; w != nil; w = cwg.queue.head {
		cwg.queue.remove(w)
		w.canceled = true
		w.cond.Signal()
	}
	cwg.settle()
	cwg.queue.mu.Unlock()
}

// Working returns a snapshot of the number of occupied worker slots.
func (cwg *ControlWaitGroup) Working() int { return int(cwg.state.Load() >> cwgShift) }

// Waiting returns a snapshot of the number of pending Do calls.
func (cwg *ControlWaitGroup) Waiting() int {
	cwg.queue.mu.Lock()
	n := cwg.queue.n
	cwg.queue.mu.Unlock()
	return n
}

// Wait waits for all admitted tasks and pending Do calls to finish.
// Submission to an empty group must precede Wait. Before reusing a group,
// all previous Wait calls must have returned.
func (cwg *ControlWaitGroup) Wait() {
	if s := cwg.state.Load(); s>>cwgShift != 0 || s&cwgQueued != 0 {
		cwg.wait(context.Background())
	}
}

// WaitContext is like Wait, but stops waiting when ctx is done. It returns
// nil once the group is idle, or ctx.Err() if ctx is done first.
func (cwg *ControlWaitGroup) WaitContext(ctx context.Context) error {
	if s := cwg.state.Load(); s>>cwgShift == 0 && s&cwgQueued == 0 {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return cwg.wait(ctx)
}

func (cwg *ControlWaitGroup) wait(ctx context.Context) error {
	cwg.queue.mu.Lock()
	for {
		s := cwg.state.Load()
		if s>>cwgShift == 0 && cwg.queue.head == nil {
			cwg.queue.mu.Unlock()
			return nil
		}
		if s&cwgWaiting != 0 || cwg.state.CompareAndSwap(s, s|cwgWaiting) {
			break
		}
	}
	w := cwg.queue.take()
	cwg.waiters.pushBack(w)
	c := w.arm(0, ctx)
	for !w.granted && !w.canceled {
		w.cond.Wait()
	}
	w.disarm(c)
	idle := w.granted
	cwg.queue.keep(w)
	cwg.queue.mu.Unlock()
	if idle {
		return nil
	}
	return ctx.Err()
}
