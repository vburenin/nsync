// Frozen pre-optimization implementation. Used only by benchmark tests.
package baseline

import "sync"

// ControlWaitGroup runs tasks with a limit on the number of concurrent goroutines.
// Use NewControlWaitGroup to initialize it. A ControlWaitGroup must not be copied
// after first use.
type ControlWaitGroup struct {
	sem     *Semaphore
	wg      sync.WaitGroup
	mu      sync.Mutex
	waiting int
	abort   bool
	done    chan struct{}
}

// NewControlWaitGroup returns a ControlWaitGroup with the given concurrency limit.
// It panics if poolSize is not positive.
func NewControlWaitGroup(poolSize int) *ControlWaitGroup {
	return &ControlWaitGroup{
		sem:  NewSemaphore(poolSize),
		done: make(chan struct{}),
	}
}

// Do waits for a free slot and starts userFunc in a new goroutine, returning true.
// If the group is aborted before the task is admitted, Do returns false without
// running userFunc. Pass arguments to userFunc through a closure.
// When the group is empty, Do must be called before Wait.
func (cwg *ControlWaitGroup) Do(userFunc func()) bool {
	cwg.mu.Lock()
	if cwg.abort {
		cwg.mu.Unlock()
		return false
	}
	cwg.wg.Add(1)
	cwg.waiting++
	cwg.mu.Unlock()

	acquired := false
	select {
	case cwg.sem.sch <- struct{}{}:
		acquired = true
	case <-cwg.done:
	}

	cwg.mu.Lock()
	cwg.waiting--

	if cwg.abort {
		if acquired {
			cwg.sem.Release()
		}
		cwg.wg.Done()
		cwg.mu.Unlock()
		return false
	}

	cwg.mu.Unlock()

	go func() {
		defer cwg.wg.Done()
		defer cwg.sem.Release()
		userFunc()
	}()
	return true
}

// Abort unblocks pending Do calls and prevents future tasks from being admitted.
// Tasks already admitted may continue running; use Wait to wait for them.
// Abort is permanent and may be called more than once.
func (cwg *ControlWaitGroup) Abort() {
	cwg.mu.Lock()
	if !cwg.abort {
		cwg.abort = true
		close(cwg.done)
	}
	cwg.mu.Unlock()
}

// Working returns the number of occupied worker slots. This is a snapshot
// intended for monitoring, not synchronization.
func (cwg *ControlWaitGroup) Working() int {
	return cwg.sem.Value()
}

// Waiting returns the number of Do calls waiting for a worker slot. This is a
// snapshot intended for monitoring, not synchronization.
func (cwg *ControlWaitGroup) Waiting() int {
	cwg.mu.Lock()
	w := cwg.waiting
	cwg.mu.Unlock()
	return w
}

// Wait blocks until all admitted tasks and pending Do calls finish.
// As with sync.WaitGroup, submission to an empty group must precede Wait.
// When reusing a group, all previous Wait calls must return before new Do calls.
func (cwg *ControlWaitGroup) Wait() {
	cwg.wg.Wait()
}
