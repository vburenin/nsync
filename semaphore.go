package nsync

import (
	"context"
	"time"
)

// Semaphore limits concurrent acquisitions. Use NewSemaphore to initialize it.
// A Semaphore must not be copied after first use.
type Semaphore struct {
	lock  lockState
	limit lockWord // immutable after construction
	// gate makes blocked or colliding acquirers take turns, so the permit
	// count is not retried by every goroutine at once. Its holder may wait
	// for a permit; releases never take it.
	gate lockState
}

// NewSemaphore creates a semaphore. It panics unless value is positive.
// Capacities above 2^58-1, or 2^26-1 (67,108,863) on 32-bit platforms, are
// reduced to that maximum.
func NewSemaphore(value int) *Semaphore {
	if value <= 0 {
		panic("nsync: semaphore capacity must be positive")
	}
	return &Semaphore{limit: min(lockWord(value), maxPermits)}
}

// Acquire acquires a slot, blocking when all slots are occupied.
func (s *Semaphore) Acquire() {
	if st := s.lock.state.Load(); st >= s.limit<<stateUnitShift || !s.lock.state.CompareAndSwap(st, st+stateUnit) {
		s.acquireSlow()
	}
}

//go:noinline
func (s *Semaphore) acquireSlow() {
	if s.limit == 1 {
		// A binary semaphore is a mutex; its queue handles contention.
		s.lock.acquireSlow(1, 0, nil)
		return
	}
	s.gate.lock()
	if !s.lock.tryAcquire(s.limit) {
		s.lock.acquireSlow(s.limit, 0, nil)
	}
	s.gate.unlock()
}

// Release releases a slot. It panics if no slot is occupied.
func (s *Semaphore) Release() {
	// A wait-free decrement: releases never retry on a contended count.
	if s.lock.state.Add(^(stateUnit-1))&(stateNegative|stateRetired|stateStarving|stateQueued) != 0 {
		s.releaseSlow()
	}
}

//go:noinline
func (s *Semaphore) releaseSlow() {
	if s.lock.state.Load()&stateNegative != 0 {
		// Undo the decrement. Acquirers that saw the negative word queued
		// meanwhile; the permit they wait for is free again. While another
		// misused release is still being undone, the word stays negative and
		// that release wakes them.
		if ns := s.lock.state.Add(stateUnit); ns&stateNegative == 0 && ns&(stateQueued|stateStarving) != 0 {
			s.lock.releaseAdded(s.limit)
		}
		panic("nsync: release without acquisition")
	}
	s.lock.releaseAdded(s.limit)
}

// TryAcquire acquires an available slot without waiting.
func (s *Semaphore) TryAcquire() bool { return s.lock.tryAcquire(s.limit) }

// TryAcquireTimeout tries immediately, then waits up to d for a slot.
// A non-positive duration is equivalent to TryAcquire.
func (s *Semaphore) TryAcquireTimeout(d time.Duration) bool {
	return s.lock.tryAcquire(s.limit) || d > 0 && s.acquireWait(d, nil)
}

// AcquireContext acquires a slot, blocking until one is available or ctx is
// done. It returns nil once a slot is acquired, or ctx.Err() without acquiring
// one. If ctx is already done, AcquireContext does not try to acquire.
func (s *Semaphore) AcquireContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.lock.tryAcquire(s.limit) || s.acquireWait(0, ctx) {
		return nil
	}
	return ctx.Err()
}

// acquireWait waits for a slot as Acquire does, for at most a positive
// timeout or until ctx is done.
func (s *Semaphore) acquireWait(timeout time.Duration, ctx context.Context) bool {
	if s.limit == 1 {
		return s.lock.acquireSlow(1, timeout, ctx) == acquired
	}
	var start time.Time
	if timeout > 0 {
		start = time.Now()
	}
	if !s.gate.state.CompareAndSwap(0, stateUnit) && s.gate.acquireSlow(1, timeout, ctx) != acquired {
		return false
	}
	ok := s.lock.tryAcquire(s.limit)
	if !ok {
		if timeout > 0 {
			timeout -= time.Since(start)
		}
		ok = (timeout > 0 || ctx != nil) && s.lock.acquireSlow(s.limit, timeout, ctx) == acquired
	}
	s.gate.unlock()
	return ok
}

// Value returns a snapshot of the number of occupied slots.
func (s *Semaphore) Value() int { return int(permits(s.lock.state.Load())) }
