// Frozen pre-optimization implementation. Used only by benchmark tests.
package baseline

import "sync"

// OnceMutex is a mutex that can be locked only once.
// Lock operation returns true if mutex has been successfully locked.
// Any other concurrent attempts will block until mutex is unlocked.
// However, any other attempts to grab a lock will return false.
// The zero value is ready to use. A OnceMutex must not be copied after first use.
type OnceMutex struct {
	mu   sync.Mutex
	used bool
}

// NewOnceMutex returns an instance of OnceMutex.
func NewOnceMutex() *OnceMutex {
	return &OnceMutex{}
}

// Lock tries to acquire lock.
func (om *OnceMutex) Lock() bool {
	om.mu.Lock()
	if om.used {
		om.mu.Unlock()
		return false
	}
	return true
}

// Unlock tries to release a lock.
func (om *OnceMutex) Unlock() {
	om.used = true
	om.mu.Unlock()
}

// NamedOnceMutex is a map of dynamically created mutexes by provided id.
// First attempt to lock by id will create a new mutex and acquire a lock.
// All other concurrent attempts will block waiting mutex to be unlocked for the same id.
// Once mutex unlocked, all other lock attempts will return false for the same instance of mutex.
// Unlocked mutex is discarded. Next attempt to acquire a lock for the same id will succeed.
// Such behaviour may be used to refresh a local cache of data identified by some key avoiding
// concurrent request to receive a refreshed value for the same key.
// Keys must be comparable and equal to themselves, as with ordinary map keys
// used for lookup (for example, do not use floating-point NaN keys).
// The zero value is ready to use. A NamedOnceMutex must not be copied after first use.
type NamedOnceMutex struct {
	lockMap map[any]*OnceMutex
	mutex   sync.Mutex
}

// NewNamedOnceMutex returns an instance of NamedOnceMutex.
func NewNamedOnceMutex() *NamedOnceMutex {
	return &NamedOnceMutex{
		lockMap: make(map[any]*OnceMutex),
	}
}

// Lock try to acquire a lock for provided id. If attempt is successful, true is returned
// If lock is already acquired by something else it will block until mutex is unlocked returning false.
func (nom *NamedOnceMutex) Lock(useMutexKey any) bool {
	m, created := nom.loadOrCreate(useMutexKey)
	if created {
		return true
	}
	return m.Lock()
}

// loadOrCreate releases the map mutex even if an invalid key causes a panic.
// A newly created OnceMutex is locked before it becomes visible to other callers.
func (nom *NamedOnceMutex) loadOrCreate(key any) (*OnceMutex, bool) {
	nom.mutex.Lock()
	defer nom.mutex.Unlock()
	m, ok := nom.lockMap[key]
	if ok {
		return m, false
	}

	if nom.lockMap == nil {
		nom.lockMap = make(map[any]*OnceMutex)
	}
	m = &OnceMutex{}
	m.Lock()
	nom.lockMap[key] = m
	return m, true
}

// Unlock unlocks the locked mutex. Used mutex will be discarded.
// Unlock does nothing if the key has no active mutex.
func (nom *NamedOnceMutex) Unlock(useMutexKey any) {
	nom.mutex.Lock()
	defer nom.mutex.Unlock()
	m, ok := nom.lockMap[useMutexKey]
	if ok {
		m.Unlock()
		delete(nom.lockMap, useMutexKey)
	}
}
