package nsync

import (
	"sync"
	"sync/atomic"
)

// ControlWaitGroup runs tasks with a limit on concurrent goroutines.
// Use NewControlWaitGroup to initialize it. A ControlWaitGroup must not be
// copied after first use.
type ControlWaitGroup struct {
	abort    atomic.Bool
	mu       sync.Mutex
	cond     sync.Cond
	finished *sync.Cond
	limit    int
	working  int
	waiting  int
}

// NewControlWaitGroup creates a group. It panics unless poolSize is positive.
func NewControlWaitGroup(poolSize int) *ControlWaitGroup {
	if poolSize <= 0 {
		panic("nsync: worker limit must be positive")
	}
	cwg := &ControlWaitGroup{limit: poolSize}
	cwg.cond.L = &cwg.mu
	return cwg
}

// Do waits for a free slot and starts userFunc in a new goroutine, returning true.
// If the group is aborted before admission, Do returns false without running
// userFunc. Submission to an empty group must precede Wait.
func (cwg *ControlWaitGroup) Do(userFunc func()) bool {
	if cwg.abort.Load() {
		return false
	}
	cwg.mu.Lock()
	if cwg.abort.Load() {
		cwg.mu.Unlock()
		return false
	}
	if cwg.working == cwg.limit {
		cwg.waiting++
		for cwg.working == cwg.limit && !cwg.abort.Load() {
			cwg.cond.Wait()
		}
		cwg.waiting--
		if cwg.abort.Load() {
			cwg.notifyFinished()
			cwg.mu.Unlock()
			return false
		}
	}
	cwg.working++
	cwg.mu.Unlock()
	go func() {
		defer cwg.finish()
		userFunc()
	}()
	return true
}

func (cwg *ControlWaitGroup) finish() {
	cwg.mu.Lock()
	cwg.working--
	if cwg.waiting != 0 {
		cwg.cond.Signal()
	}
	cwg.notifyFinished()
	cwg.mu.Unlock()
}

// Abort unblocks pending Do calls and permanently rejects new tasks.
// Already admitted tasks may continue running; use Wait to wait for them.
// Repeated calls are safe.
func (cwg *ControlWaitGroup) Abort() {
	cwg.mu.Lock()
	cwg.abort.Store(true)
	cwg.cond.Broadcast()
	cwg.mu.Unlock()
}

// Working returns a snapshot of the number of occupied worker slots.
func (cwg *ControlWaitGroup) Working() int {
	cwg.mu.Lock()
	n := cwg.working
	cwg.mu.Unlock()
	return n
}

// Waiting returns a snapshot of the number of pending Do calls.
func (cwg *ControlWaitGroup) Waiting() int {
	cwg.mu.Lock()
	n := cwg.waiting
	cwg.mu.Unlock()
	return n
}

// Wait waits for all admitted tasks and pending Do calls to finish.
// Submission to an empty group must precede Wait. Before reusing a group,
// all previous Wait calls must have returned.
func (cwg *ControlWaitGroup) Wait() {
	cwg.mu.Lock()
	for cwg.working != 0 || cwg.waiting != 0 {
		if cwg.finished == nil {
			cwg.finished = sync.NewCond(&cwg.mu)
		}
		cwg.finished.Wait()
	}
	cwg.mu.Unlock()
}

// Called with mu held. Completion waiters use a separate condition variable
// so a worker-slot signal cannot accidentally wake the wrong class of waiter.
func (cwg *ControlWaitGroup) notifyFinished() {
	if cwg.working == 0 && cwg.waiting == 0 && cwg.finished != nil {
		cwg.finished.Broadcast()
	}
}
