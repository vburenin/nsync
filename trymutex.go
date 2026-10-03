package nsync

import (
	"context"
	"time"
)

// TryMutex provides blocking, nonblocking, timed, and context-aware lock
// acquisition. Use NewTryMutex to initialize it. Copies refer to the same
// underlying lock.
type TryMutex struct{ state *lockState }

// NewTryMutex creates an unlocked mutex.
func NewTryMutex() *TryMutex {
	// Allocate the public handle and its shared state together. Value copies
	// still refer to the same state, without a second heap allocation.
	m := new(struct {
		handle TryMutex
		state  lockState
	})
	m.handle.state = &m.state
	return &m.handle
}

// Lock acquires the mutex, blocking if it is already locked.
func (tm TryMutex) Lock() { tm.state.lock() }

// TryLock acquires the mutex without waiting, returning true on success.
func (tm TryMutex) TryLock() bool { return tm.state.tryLock() }

// TryLockTimeout tries immediately, then waits up to timeout for the mutex.
// A non-positive timeout is equivalent to TryLock.
func (tm TryMutex) TryLockTimeout(timeout time.Duration) bool { return tm.state.lockTimeout(timeout) }

// LockContext acquires the mutex, blocking until it is available or ctx is
// done. It returns nil once the mutex is acquired, or ctx.Err() without
// acquiring it. If ctx is already done, LockContext does not try to acquire.
func (tm TryMutex) LockContext(ctx context.Context) error { return tm.state.lockContext(ctx) }

// Unlock releases the mutex. It panics if the mutex is unlocked.
func (tm TryMutex) Unlock() { tm.state.unlock() }
