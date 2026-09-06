package nsync

import (
	"sync"
	"sync/atomic"
)

// SyncFlag implements a boolean flag that can be set or unset atomically.
// During set/unset SyncFlag locks the mutex, so if anything needs to prevent
// a flag from being set/unset should acquire a lock.
// The zero value is unset. A SyncFlag must not be copied after first use.
// Set and Unset must not be called while the caller holds the embedded mutex.
type SyncFlag struct {
	sync.Mutex
	flag atomic.Bool
}

// Set locks the mutex, sets the flag and unlocks the mutex.
func (bf *SyncFlag) Set() {
	bf.Lock()
	bf.flag.Store(true)
	bf.Unlock()
}

// Unset locks the mutex, resets the flag and unlocks the mutex.
func (bf *SyncFlag) Unset() {
	bf.Lock()
	bf.flag.Store(false)
	bf.Unlock()
}

// IsSet atomically checks if flag is set.
func (bf *SyncFlag) IsSet() bool {
	return bf.flag.Load()
}

// IsUnset atomically checks if flag is unset.
func (bf *SyncFlag) IsUnset() bool {
	return !bf.flag.Load()
}
