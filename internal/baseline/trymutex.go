// Frozen pre-optimization implementation. Used only by benchmark tests.
package baseline

import "time"

// TryMutex is a channel-based mutex with nonblocking and timed acquisition.
// Use NewTryMutex to initialize it.
type TryMutex struct {
	c chan struct{}
}

// NewTryMutex makes a new TryMutex instance.
func NewTryMutex() *TryMutex {
	return &TryMutex{
		c: make(chan struct{}, 1),
	}
}

// Lock acquires the lock.
func (tm TryMutex) Lock() {
	tm.c <- struct{}{}
}

// TryLock tries to acquire the lock, returning true on success.
func (tm TryMutex) TryLock() bool {
	select {
	case tm.c <- struct{}{}:
		return true
	default:
		return false
	}
}

// TryLockTimeout tries to acquire the lock, returning true on success.
// Attempt to acquire the lock will timeout after the caller defined interval.
// A non-positive timeout is equivalent to TryLock.
func (tm TryMutex) TryLockTimeout(timeout time.Duration) bool {
	return acquireTimeout(tm.c, timeout)
}

// Unlock releases the lock. If lock hasn't been acquired
// function will panic.
func (tm TryMutex) Unlock() {
	select {
	case <-tm.c:
	default:
		panic("nsync: unlock of unlocked TryMutex")
	}
}
