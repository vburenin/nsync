package nsync

import (
	"sync"
	"time"
)

// gate serializes short counter updates with the adaptive mutex.
// Waiting for a permit releases that mutex and parks on a condition variable.
type gate struct {
	used    int64
	limit   int64 // Immutable after construction; reads do not require mu.
	mu      mutexState
	waiters int
	cond    *sync.Cond
}

func (g *gate) tryAcquire() bool {
	g.mu.lock()
	used := g.used
	if used == g.limit {
		g.mu.unlock()
		return false
	}
	g.used = used + 1
	g.mu.unlock()
	return true
}

func (g *gate) acquire() {
	g.mu.lock()
	for g.used == g.limit {
		if g.cond == nil {
			g.cond = sync.NewCond(&g.mu)
		}
		g.waiters++
		g.cond.Wait()
		g.waiters--
	}
	g.used = g.used + 1
	g.mu.unlock()
}

func (g *gate) acquireTimeout(timeout time.Duration) bool {
	if g.tryAcquire() {
		return true
	}
	if timeout <= 0 {
		return false
	}
	return g.acquireSlow(timeout)
}

func (g *gate) acquireSlow(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	g.mu.lock()
	defer g.mu.unlock()
	if g.used < g.limit {
		g.used = g.used + 1
		return true
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return false
	}
	if g.cond == nil {
		g.cond = sync.NewCond(&g.mu)
	}
	g.waiters++
	defer func() { g.waiters-- }()
	w, timer := startTimedWait(g.cond, deadline)
	if w == nil {
		return false
	}
	defer w.stop(timer)
	for {
		g.cond.Wait()
		if w.expired {
			return false
		}
		if g.used < g.limit {
			g.used = g.used + 1
			return true
		}
	}
}

func (g *gate) release() {
	g.mu.lock()
	used := g.used
	if used == 0 {
		g.mu.unlock()
		panic("nsync: release without acquisition")
	}
	g.used = used - 1
	if g.waiters != 0 {
		g.cond.Signal()
	}
	g.mu.unlock()
}
