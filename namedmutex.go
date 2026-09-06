package nsync

import (
	"hash/maphash"
	"sync/atomic"
	"time"
	"unsafe"
)

// NamedMutex provides independent locks by name. The zero value is ready to use.
// A NamedMutex must not be copied after first use. Locks are retained for the
// lifetime of the NamedMutex, so the set of names should be bounded.
type NamedMutex struct {
	first atomic.Pointer[namedPrimaryEntry]
	table atomic.Pointer[namedMutexTable]
}

type namedMutexTable [64]atomic.Pointer[namedMutexShard]

// The name is read by every lookup, so keep it on a separate cache
// line from the lock that its owner modifies.
type namedPrimaryEntry struct {
	name  string
	_     [cacheLineSize - unsafe.Sizeof("")]byte
	mutex mutexState
	_     [(cacheLineSize - unsafe.Sizeof(mutexState{})%cacheLineSize) % cacheLineSize]byte
}

type namedMutexShard struct {
	first atomic.Pointer[namedMutexEntry]
	mu    mutexState
	locks map[string]*mutexState
}

// The first key in each shard has a read-only lookup slot. Its lock occupies
// its own cache line, avoiding writes shared with an independent hot key.
type namedMutexEntryData struct {
	name  string
	mutex mutexState
}

type namedMutexEntry struct {
	_ [(cacheLineSize - unsafe.Sizeof(namedMutexEntryData{})%cacheLineSize) % cacheLineSize]byte
	namedMutexEntryData
}

var namedMutexSeed = maphash.MakeSeed()

// NewNamedMutex creates a named mutex.
func NewNamedMutex() *NamedMutex { return &NamedMutex{} }

func (nm *NamedMutex) lookup(name string, create bool) *mutexState {
	// A repeated name avoids hashing, including long strings. Checking the
	// suffix first rejects common-prefix names without comparing every byte.
	primary := nm.first.Load()
	if primary == nil && create {
		primary = &namedPrimaryEntry{name: name}
		if !nm.first.CompareAndSwap(nil, primary) {
			primary = nm.first.Load()
		}
	}
	if primary != nil && len(primary.name) == len(name) &&
		(len(name) <= 32 || primary.name[len(name)-8:] == name[len(name)-8:]) &&
		primary.name == name {
		return &primary.mutex
	}
	table := nm.table.Load()
	if table == nil {
		if !create {
			return nil
		}
		table = new(namedMutexTable)
		if !nm.table.CompareAndSwap(nil, table) {
			table = nm.table.Load()
		}
	}
	cell := &table[maphash.String(namedMutexSeed, name)&63]
	sh := cell.Load()
	if sh == nil {
		if !create {
			return nil
		}
		sh = new(namedMutexShard)
		if !cell.CompareAndSwap(nil, sh) {
			sh = cell.Load()
		}
	}
	first := sh.first.Load()
	if first == nil {
		if !create {
			return nil
		}
		first = &namedMutexEntry{namedMutexEntryData: namedMutexEntryData{name: name}}
		if !sh.first.CompareAndSwap(nil, first) {
			first = sh.first.Load()
		}
	}
	if first.name == name {
		return &first.mutex
	}
	sh.mu.lock()
	m := sh.locks[name]
	if m == nil && create {
		m = new(mutexState)
		if sh.locks == nil {
			sh.locks = make(map[string]*mutexState)
		}
		sh.locks[name] = m
	}
	sh.mu.unlock()
	return m
}

// Lock acquires the named lock, creating it if necessary.
func (nm *NamedMutex) Lock(name string) { nm.lookup(name, true).lock() }

// TryLock tries to acquire the named lock without waiting.
func (nm *NamedMutex) TryLock(name string) bool { return nm.lookup(name, true).tryLock() }

// TryLockTimeout tries immediately, then waits up to timeout for the named lock.
// A non-positive timeout is equivalent to TryLock.
func (nm *NamedMutex) TryLockTimeout(name string, timeout time.Duration) bool {
	return nm.lookup(name, true).lockTimeout(timeout)
}

// Unlock releases the named lock. It panics if the name is unknown or unlocked.
func (nm *NamedMutex) Unlock(name string) {
	m := nm.lookup(name, false)
	if m == nil {
		panic("nsync: unknown mutex: " + name)
	}
	m.unlock()
}
