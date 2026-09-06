package nsync

import (
	"hash/maphash"
	"sync"
	"sync/atomic"
	"unsafe"
)

// OnceMutex admits one operation. The first Lock returns true; other calls
// wait for Unlock and return false. Its zero value is ready to use. It must not
// be copied after first use.
type OnceMutex struct {
	state atomic.Uint32
	wait  atomic.Pointer[onceWait]
}

type onceWait struct {
	mu   sync.Mutex
	cond sync.Cond
}

// NewOnceMutex creates a once mutex.
func NewOnceMutex() *OnceMutex { return &OnceMutex{} }

// Lock starts the operation or waits for the successful caller's Unlock.
func (om *OnceMutex) Lock() bool {
	if om.state.Load() == 2 {
		return false
	}
	if om.state.CompareAndSwap(0, 1) {
		return true
	}
	om.waitSlow()
	return false
}

func (om *OnceMutex) waitSlow() {
	if om.state.Load() == 2 {
		return
	}
	w := om.wait.Load()
	if w == nil {
		w = new(onceWait)
		w.cond.L = &w.mu
		if !om.wait.CompareAndSwap(nil, w) {
			w = om.wait.Load()
		}
	}
	w.mu.Lock()
	for om.state.Load() != 2 {
		w.cond.Wait()
	}
	w.mu.Unlock()
}

// Unlock completes the operation and publishes its writes to waiting callers.
// It panics unless Lock has succeeded and Unlock has not been called before.
func (om *OnceMutex) Unlock() {
	if !om.state.CompareAndSwap(1, 2) {
		panic("nsync: unlock of inactive OnceMutex")
	}
	if w := om.wait.Load(); w != nil {
		w.mu.Lock()
		w.cond.Broadcast()
		w.mu.Unlock()
	}
}

// NamedOnceMutex combines overlapping operations for the same comparable key.
// One caller's Lock returns true; concurrent callers wait for its Unlock and
// return false. After Unlock, a new call may start another operation for the key.
// The zero value is ready to use and must not be copied after first use.
// Keys must be comparable and equal to themselves (NaN keys are unsupported).
type NamedOnceMutex struct {
	namedOnceIndex
	_            [cacheLineSize - unsafe.Sizeof(namedOnceIndex{})]byte
	primaryShard namedOnceShard
}

type namedOnceIndex struct {
	firstBucket atomic.Uint32
	table       atomic.Pointer[namedOnceTable]
}

type namedOnceTable [namedOnceShardCount]atomic.Pointer[namedOnceShard]

const namedOnceShardCount = 64

var namedOnceSeed = maphash.MakeSeed()

type namedOnceShardData struct {
	mu         mutexState
	primaryKey any
	primary    *namedOnceEntry
	locks      map[any]*namedOnceEntry
	extraCount int
	spare      *namedOnceEntry
}

type namedOnceShard struct {
	_ [(cacheLineSize - unsafe.Sizeof(namedOnceShardData{})%cacheLineSize) % cacheLineSize]byte
	namedOnceShardData
}

type namedOnceEntry struct {
	once   OnceMutex
	shared bool
}

// NewNamedOnceMutex creates a named once mutex.
func NewNamedOnceMutex() *NamedOnceMutex { return &NamedOnceMutex{} }

func (nom *NamedOnceMutex) shard(key any, create bool) *namedOnceShard {
	// Hashing validates comparability before any mutex is acquired.
	index := maphash.Comparable(namedOnceSeed, key) & (namedOnceShardCount - 1)
	first := nom.firstBucket.Load()
	if first == 0 && create {
		if nom.firstBucket.CompareAndSwap(0, uint32(index)+1) {
			return &nom.primaryShard
		}
		first = nom.firstBucket.Load()
	}
	if first == uint32(index)+1 {
		return &nom.primaryShard
	}
	table := nom.table.Load()
	if table == nil {
		if !create {
			return nil
		}
		table = new(namedOnceTable)
		if !nom.table.CompareAndSwap(nil, table) {
			table = nom.table.Load()
		}
	}
	cell := &table[index]
	if sh := cell.Load(); sh != nil {
		return sh
	}
	if !create {
		return nil
	}
	sh := new(namedOnceShard)
	if cell.CompareAndSwap(nil, sh) {
		return sh
	}
	return cell.Load()
}

// Lock starts an operation for key, or waits for the current operation to finish.
// Only callers receiving true should call Unlock.
func (nom *NamedOnceMutex) Lock(key any) bool {
	sh := nom.shard(key, true)
	sh.mu.Lock()
	entry := sh.primary
	if entry == nil || sh.primaryKey != key {
		entry = nil
		if sh.extraCount != 0 {
			entry = sh.locks[key]
		}
	}
	if entry != nil {
		// Mark before dropping the map lock. Unlock must not recycle this
		// entry even if this caller has not yet entered OnceMutex.Lock.
		entry.shared = true
		sh.mu.Unlock()
		return entry.once.Lock()
	}
	entry = sh.spare
	if entry == nil {
		entry = new(namedOnceEntry)
	} else {
		sh.spare = nil
	}
	entry.once.state.Store(1)
	if sh.primary == nil {
		sh.primaryKey, sh.primary = key, entry
	} else {
		if sh.locks == nil {
			sh.locks = make(map[any]*namedOnceEntry)
		}
		sh.locks[key] = entry
		sh.extraCount++
	}
	sh.mu.Unlock()
	return true
}

// Unlock completes the active operation for key. Unknown keys are ignored.
// Only unshared entries can be recycled; a shared entry remains alive through
// the references held by its waiters, independently of later operations.
func (nom *NamedOnceMutex) Unlock(key any) {
	sh := nom.shard(key, false)
	if sh == nil {
		return
	}
	sh.mu.Lock()
	var entry *namedOnceEntry
	if sh.primary != nil && sh.primaryKey == key {
		entry = sh.primary
		sh.primaryKey, sh.primary = nil, nil
	} else if sh.extraCount != 0 {
		entry = sh.locks[key]
		if entry != nil {
			delete(sh.locks, key)
			sh.extraCount--
		}
	}
	if entry != nil {
		if entry.shared {
			entry.once.Unlock()
		} else {
			// No waiter has ever received this entry. Its completed state can
			// be published directly, without a compare-and-swap or wakeup.
			entry.once.state.Store(2)
		}
		if !entry.shared && sh.spare == nil {
			sh.spare = entry
		}
	}
	sh.mu.Unlock()
}
