package nsync

import (
	"sync"
	"sync/atomic"
	"time"
)

// mutexState uses four states: unlocked, locked, locked with waiters, and
// reserved for a woken waiter. Adaptive handoff prevents sustained barging.
// The uncontended path needs one atomic operation to acquire and one to release.
// Waiters park through sync.Cond, whose runtime implementation is portable.
type mutexState struct {
	state   atomic.Int32
	mu      sync.Mutex
	fair    bool
	waiters int
	cond    sync.Cond
}

// A shorter threshold reduced long waits in the contention benchmarks without
// the throughput cost of handing off on every acquisition. It is a scheduling
// heuristic, not a bound on how long Lock can take.
const handoffThreshold = 100 * time.Microsecond

// Lock and Unlock implement sync.Locker for condition variables.
func (m *mutexState) Lock()   { m.lock() }
func (m *mutexState) Unlock() { m.unlock() }

func (m *mutexState) lock() {
	if !m.state.CompareAndSwap(0, 1) {
		m.lockSlow(0, false)
	}
}

func (m *mutexState) tryLock() bool {
	return m.state.Load() == 0 && m.state.CompareAndSwap(0, 1)
}

func (m *mutexState) lockTimeout(timeout time.Duration) bool {
	if m.state.CompareAndSwap(0, 1) {
		return true
	}
	if timeout <= 0 {
		return false
	}
	return m.lockSlow(timeout, true)
}

func (m *mutexState) lockSlow(timeout time.Duration, timed bool) (acquired bool) {
	start := time.Now()
	var deadline time.Time
	if timed {
		deadline = start.Add(timeout)
	}
	if !timed && m.spin() {
		return true
	}
	m.mu.Lock()
	if m.cond.L == nil {
		m.cond.L = &m.mu
	}
	m.waiters++
	defer func() {
		m.waiters--
		if m.waiters == 0 {
			m.fair = false
			m.state.CompareAndSwap(2, 1)
			if !acquired {
				m.state.CompareAndSwap(3, 0)
			}
		} else if !acquired && m.state.Load() == 3 {
			// A timed-out waiter must pass a reserved handoff onward.
			m.cond.Signal()
		}
		m.mu.Unlock()
	}()
	if m.acquireQueued(false) {
		return true
	}
	if timed {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false
		}
		return m.waitTimed(start, deadline)
	}
	for {
		m.cond.Wait()
		if m.acquireQueued(true) {
			m.fair = time.Since(start) >= handoffThreshold
			return true
		}
		if time.Since(start) >= handoffThreshold {
			m.fair = true
		}
	}
}

// acquireQueued is called with mu held. Only a caller returning from Cond.Wait
// may consume a handoff; a new arrival must join the wait queue instead.
func (m *mutexState) acquireQueued(woken bool) bool {
	state := m.state.Swap(2)
	if state == 3 && !woken {
		m.state.Store(3)
	}
	return state == 0 || state == 3 && woken
}

// waitTimed is separate so ordinary blocking Lock does not allocate a timer's
// captured state. mu is held on entry and on return.
func (m *mutexState) waitTimed(start, deadline time.Time) bool {
	// Recheck immediately before allocating a timer: a release may have raced
	// with registration, or a very short deadline may already have elapsed.
	if m.acquireQueued(false) {
		return true
	}
	timeout := time.Until(deadline)
	if timeout <= 0 {
		return false
	}
	w, timer := startTimedWait(&m.cond, deadline)
	if w == nil {
		return false
	}
	defer w.stop(timer)
	for {
		m.cond.Wait()
		if w.expired {
			return false
		}
		if m.acquireQueued(true) {
			m.fair = time.Since(start) >= handoffThreshold
			return true
		}
		if time.Since(start) >= handoffThreshold {
			m.fair = true
		}
	}
}

func (m *mutexState) unlock() {
	if !m.state.CompareAndSwap(1, 0) {
		m.unlockSlow()
	}
}

func (m *mutexState) unlockSlow() {
	m.mu.Lock()
	// The last timed waiter may have changed 2 back to 1 after our fast
	// compare-and-swap failed, but before we acquired the queue mutex.
	if state := m.state.Load(); state != 1 && state != 2 {
		m.mu.Unlock()
		panic("nsync: unlock of unlocked mutex")
	}
	if m.fair && m.waiters != 0 {
		m.state.Store(3)
	} else {
		m.state.Store(0)
	}
	m.cond.Signal()
	m.mu.Unlock()
}
