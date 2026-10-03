package nsync

import (
	"context"
	"hash/maphash"
	"sync/atomic"
	"unsafe"
)

// OnceMutex admits one operation. The first Lock returns true; other calls
// wait for Unlock and return false. Its zero value is ready to use. It must not
// be copied after first use.
type OnceMutex struct {
	state atomic.Uint32
	queue atomic.Pointer[waitQueue]
}

const (
	onceIdle    uint32 = 0
	onceRunning uint32 = 1
	onceDone    uint32 = 2
	onceWaiting uint32 = 4 // waiters may be queued: Unlock must wake them
)

// NewOnceMutex creates a once mutex.
func NewOnceMutex() *OnceMutex { return &OnceMutex{} }

// Lock starts the operation or waits for the successful caller's Unlock.
func (om *OnceMutex) Lock() bool {
	return om.state.Load() != onceDone && om.lockSlow()
}

//go:noinline
func (om *OnceMutex) lockSlow() bool {
	if om.state.CompareAndSwap(onceIdle, onceRunning) {
		return true
	}
	leader, _ := om.wait(context.Background())
	return leader
}

// LockContext is like Lock, but stops waiting when ctx is done. It returns
// (true, nil) if the caller must perform the operation and then call Unlock,
// (false, nil) once another caller's operation has completed, or
// (false, ctx.Err()) if ctx is done first. A done ctx never starts the
// operation, but an already completed operation is reported as such.
func (om *OnceMutex) LockContext(ctx context.Context) (bool, error) {
	if om.state.Load() == onceDone {
		return false, nil
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if om.state.CompareAndSwap(onceIdle, onceRunning) {
		return true, nil
	}
	return om.wait(ctx)
}

// wait waits for the running operation to complete, or for ctx to be done.
func (om *OnceMutex) wait(ctx context.Context) (bool, error) {
	q := om.queue.Load()
	if q == nil {
		q = new(waitQueue)
		if !om.queue.CompareAndSwap(nil, q) {
			q = om.queue.Load()
		}
	}
	q.mu.Lock()
	for {
		s := om.state.Load()
		if s == onceDone {
			q.mu.Unlock()
			return false, nil
		}
		if s == onceIdle {
			if om.state.CompareAndSwap(onceIdle, onceRunning) {
				q.mu.Unlock()
				return true, nil
			}
			continue
		}
		if s&onceWaiting != 0 || om.state.CompareAndSwap(s, s|onceWaiting) {
			break
		}
	}
	w := q.take()
	q.pushBack(w)
	c := w.arm(0, ctx)
	for !w.granted && !w.canceled {
		w.cond.Wait()
	}
	w.disarm(c)
	done := w.granted
	q.keep(w)
	q.mu.Unlock()
	if done || om.state.Load() == onceDone {
		return false, nil
	}
	return false, ctx.Err()
}

// Unlock completes the operation and publishes its writes to waiting callers.
// It panics unless Lock has succeeded and Unlock has not been called before.
func (om *OnceMutex) Unlock() {
	if !om.state.CompareAndSwap(onceRunning, onceDone) {
		om.unlockSlow()
	}
}

//go:noinline
func (om *OnceMutex) unlockSlow() {
	if !om.state.CompareAndSwap(onceRunning|onceWaiting, onceDone) {
		panic("nsync: unlock of inactive OnceMutex")
	}
	q := om.queue.Load()
	q.mu.Lock()
	q.grantAll()
	q.mu.Unlock()
}

// NamedOnceMutex combines overlapping operations for the same comparable key.
// One caller's Lock returns true; concurrent callers wait for its Unlock and
// return false. After Unlock, a new call may start another operation for the key.
// The zero value is ready to use and must not be copied after first use.
// Keys must be comparable and equal to themselves (NaN keys are unsupported).
type NamedOnceMutex struct {
	namedOnceIndex
	_            [(cacheLineSize - unsafe.Sizeof(namedOnceIndex{})%cacheLineSize) % cacheLineSize]byte
	primaryShard namedOnceShard
}

type namedOnceIndex struct {
	firstBucket atomic.Uint32
	// The key that chose firstBucket is kept when it is a string or an
	// integer: an equal key belongs to the primary shard, so lookups compare
	// it before hashing. The key is written once, before firstKind publishes
	// it; lookups read it only after loading that kind.
	firstKind atomic.Uint32
	firstInt  uint64
	firstStr  string
	table     atomic.Pointer[namedOnceTable]
}

// Kinds of the first key that lookups compare before hashing.
const (
	firstNone uint32 = iota
	firstString
	firstInt
	firstInt64
	firstUint64
)

type namedOnceTable [namedOnceShardCount]atomic.Pointer[namedOnceShard]

const namedOnceShardCount = 64

var namedOnceSeed = maphash.MakeSeed()

type namedOnceShardData struct {
	mu         lockState
	primaryKey any
	primary    *namedOnceEntry
	locks      map[any]*namedOnceEntry
	extraCount int
	spare      *namedOnceEntry // an unshared entry, still running, for reuse
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

// shard returns the shard for key. Common key types skip the generic hash,
// which looks up the hasher of the dynamic type; keys of different types are
// never equal, so each type may hash differently. Hashing panics, before any
// mutex is acquired, for keys that are not comparable.
func (nom *NamedOnceMutex) shard(key any, create bool) *namedOnceShard {
	var h uint64
	kind := firstNone
	switch k := key.(type) {
	case string:
		if nom.firstKind.Load() == firstString && nom.firstStr == k {
			return &nom.primaryShard
		}
		h, kind = maphash.String(namedOnceSeed, k), firstString
	case int:
		if nom.firstKind.Load() == firstInt && nom.firstInt == uint64(k) {
			return &nom.primaryShard
		}
		h, kind = maphash.Comparable(namedOnceSeed, k), firstInt
	case int64:
		if nom.firstKind.Load() == firstInt64 && nom.firstInt == uint64(k) {
			return &nom.primaryShard
		}
		h, kind = maphash.Comparable(namedOnceSeed, k), firstInt64
	case uint64:
		if nom.firstKind.Load() == firstUint64 && nom.firstInt == k {
			return &nom.primaryShard
		}
		h, kind = maphash.Comparable(namedOnceSeed, k), firstUint64
	default:
		h = maphash.Comparable(namedOnceSeed, key)
	}
	index := h & (namedOnceShardCount - 1)
	first := nom.firstBucket.Load()
	if first == 0 && create {
		if nom.firstBucket.CompareAndSwap(0, uint32(index)+1) {
			nom.setFirstKey(key, kind)
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

// setFirstKey records the key that chose firstBucket, if lookups can compare
// it before hashing.
func (nom *NamedOnceMutex) setFirstKey(key any, kind uint32) {
	switch k := key.(type) {
	case string:
		nom.firstStr = k
	case int:
		nom.firstInt = uint64(k)
	case int64:
		nom.firstInt = uint64(k)
	case uint64:
		nom.firstInt = k
	}
	if kind != firstNone {
		nom.firstKind.Store(kind)
	}
}

// Lock starts an operation for key, or waits for the current operation to finish.
// Only callers receiving true should call Unlock.
func (nom *NamedOnceMutex) Lock(key any) bool {
	entry := nom.join(key)
	return entry == nil || entry.once.Lock()
}

// LockContext is like Lock, but stops waiting when ctx is done. It returns
// (true, nil) if the caller must perform the operation and then call Unlock,
// (false, nil) once the operation in progress has completed, or
// (false, ctx.Err()) if ctx is done first. A done ctx never starts an operation.
func (nom *NamedOnceMutex) LockContext(ctx context.Context, key any) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	entry := nom.join(key)
	if entry == nil {
		return true, nil
	}
	return entry.once.LockContext(ctx)
}

// join starts an operation for key and returns nil, or returns the entry of
// the operation already in progress.
func (nom *NamedOnceMutex) join(key any) *namedOnceEntry {
	sh := nom.shard(key, true)
	sh.mu.lock()
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
		sh.mu.unlock()
		return entry
	}
	// A spare is still in the running state its previous leader left it in.
	entry = sh.spare
	if entry == nil {
		entry = new(namedOnceEntry)
		entry.once.state.Store(onceRunning)
	} else {
		sh.spare = nil
	}
	if sh.primary == nil {
		sh.primaryKey, sh.primary = key, entry
	} else {
		if sh.locks == nil {
			sh.locks = make(map[any]*namedOnceEntry)
		}
		sh.locks[key] = entry
		sh.extraCount++
	}
	sh.mu.unlock()
	return nil
}

// Unlock completes the active operation for key. Unknown keys are ignored.
// Only unshared entries can be recycled; a shared entry remains alive through
// the references held by its waiters, independently of later operations.
func (nom *NamedOnceMutex) Unlock(key any) {
	sh := nom.shard(key, false)
	if sh == nil {
		return
	}
	sh.mu.lock()
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
		} else if sh.spare == nil {
			// No other caller has ever received this entry, so its state is
			// unobservable: it can be reused without being marked done.
			sh.spare = entry
		}
	}
	sh.mu.unlock()
}
