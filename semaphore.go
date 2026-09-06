package nsync

import "time"

// Semaphore limits concurrent acquisitions. Use NewSemaphore to initialize it.
// A Semaphore must not be copied after first use.
type Semaphore struct{ gate gate }

// NewSemaphore creates a semaphore. It panics unless value is positive.
func NewSemaphore(value int) *Semaphore {
	if value <= 0 {
		panic("nsync: semaphore capacity must be positive")
	}
	return &Semaphore{gate: gate{limit: int64(value)}}
}

// Acquire acquires a slot, blocking when all slots are occupied.
func (s *Semaphore) Acquire() {
	if s.gate.limit == 1 {
		s.gate.mu.lock()
		return
	}
	s.gate.acquire()
}

// Release releases a slot. It panics if no slot is occupied.
func (s *Semaphore) Release() {
	if s.gate.limit == 1 {
		s.gate.mu.unlock()
		return
	}
	s.gate.release()
}

// TryAcquire acquires an available slot without waiting.
func (s *Semaphore) TryAcquire() bool {
	if s.gate.limit == 1 {
		return s.gate.mu.tryLock()
	}
	return s.gate.tryAcquire()
}

// TryAcquireTimeout tries immediately, then waits up to d for a slot.
// A non-positive duration is equivalent to TryAcquire.
func (s *Semaphore) TryAcquireTimeout(d time.Duration) bool {
	if s.gate.limit == 1 {
		return s.gate.mu.lockTimeout(d)
	}
	return s.gate.acquireTimeout(d)
}

// Value returns a snapshot of the number of occupied slots.
func (s *Semaphore) Value() int {
	if s.gate.limit == 1 {
		return int(min(s.gate.mu.state.Load(), 1))
	}
	s.gate.mu.lock()
	n := int(s.gate.used)
	s.gate.mu.unlock()
	return n
}
