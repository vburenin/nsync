// Frozen pre-optimization implementation. Used only by benchmark tests.
// Semaphore implementation that adds so necessary synchronization
// primitive into Go language. It uses built-in channel with empty struct
// so it doesn't utilize a lot of memory to buffer acquired elements.

package baseline

import "time"

// Semaphore limits the number of concurrent acquisitions using a buffered channel.
// Use NewSemaphore to initialize it.
type Semaphore struct {
	sch chan struct{}
}

// NewSemaphore returns an instance of a semaphore.
// It panics if value is not positive.
func NewSemaphore(value int) *Semaphore {
	if value <= 0 {
		panic("nsync: semaphore capacity must be positive")
	}
	return &Semaphore{
		sch: make(chan struct{}, value),
	}
}

// Acquire tries to acquire semaphore lock. If no luck it will block.
func (s *Semaphore) Acquire() {
	s.sch <- struct{}{}
}

// Release releases acquired semaphore. If semaphore is not acquired it will panic.
func (s *Semaphore) Release() {
	select {
	case <-s.sch:
	default:
		panic("No semaphore locks!")
	}
}

// TryAcquire tries to acquire semaphore. Returns true/false if success/failure accordingly.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.sch <- struct{}{}:
		return true
	default:
		return false
	}
}

// TryAcquireTimeout tries to acquire semaphore for a specified time interval.
// Returns true/false if success/failure accordingly.
// A non-positive interval is equivalent to TryAcquire.
func (s *Semaphore) TryAcquireTimeout(d time.Duration) bool {
	return acquireTimeout(s.sch, d)
}

// Value returns the number of currently acquired semaphores.
func (s *Semaphore) Value() int {
	return len(s.sch)
}
