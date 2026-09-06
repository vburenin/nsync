// Frozen pre-optimization implementation. Used only by benchmark tests.
package baseline

import (
	"sync"
	"time"
)

// NamedMutex acquires a lock based on a user defined name.
// Can be used in factories which produce singleton objects depending
// on the name. For example: queue name, user name, etc.
// Locks are based on channels, so an additional TryLock has been introduced.
// The zero value is ready to use. A NamedMutex must not be copied after first use.
// Locks are retained for the lifetime of the NamedMutex, so the set of names
// should be bounded.
type NamedMutex struct {
	mutexMap   map[string]chan struct{}
	localMutex sync.Mutex
}

// NewNamedMutex makes a new named mutex instance.
func NewNamedMutex() *NamedMutex {
	return &NamedMutex{
		mutexMap: make(map[string]chan struct{}),
	}
}

func (nm *NamedMutex) channel(name string) chan struct{} {
	nm.localMutex.Lock()
	defer nm.localMutex.Unlock()
	if nm.mutexMap == nil {
		nm.mutexMap = make(map[string]chan struct{})
	}
	mc, ok := nm.mutexMap[name]
	if !ok {
		mc = make(chan struct{}, 1)
		nm.mutexMap[name] = mc
	}
	return mc
}

// Lock acquires the lock. On the first lock attempt
// a new channel is automatically created.
func (nm *NamedMutex) Lock(name string) {
	nm.channel(name) <- struct{}{}
}

// TryLock tries to acquire the lock, returning true on success.
// On the first lock attempt a new channel is automatically created.
func (nm *NamedMutex) TryLock(name string) bool {
	select {
	case nm.channel(name) <- struct{}{}:
		return true
	default:
		return false
	}
}

// TryLockTimeout tries to acquire the lock, returning true on success.
// Attempt to acquire the lock will timeout after the caller defined interval.
// On the first lock attempt a new channel is automatically created.
// A non-positive timeout is equivalent to TryLock.
func (nm *NamedMutex) TryLockTimeout(name string, timeout time.Duration) bool {
	return acquireTimeout(nm.channel(name), timeout)
}

// Unlock releases the lock. If lock hasn't been acquired
// function will panic.
func (nm *NamedMutex) Unlock(name string) {
	nm.localMutex.Lock()
	defer nm.localMutex.Unlock()
	select {
	case <-nm.mutexMap[name]:
	default:
		panic("No named mutex acquired: " + name)
	}
}
