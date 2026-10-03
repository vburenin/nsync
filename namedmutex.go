package nsync

import (
	"context"
	"hash/maphash"
	"sync/atomic"
	"time"
	"unsafe"
)

// NamedMutex provides independent locks by name. The zero value is ready to
// use. A NamedMutex must not be copied after first use.
//
// A name's lock exists while a goroutine holds it or waits for it; it is
// discarded afterwards, so memory does not grow with the number of names ever
// used. To keep repeated use of the same names allocation-free, an instance
// retains a bounded number of idle locks: one for the first name it locked,
// and at most one cached and one spare lock in each of its 64 shards.
type NamedMutex struct {
	first atomic.Pointer[namedFirst] // never replaced or discarded once set
	table atomic.Pointer[namedTable]
}

// The first name is compared by every lookup and never changes, so keep it
// on a separate cache line from the lock that its owner writes.
type namedFirst struct {
	name string
	_    [cacheLineSize - unsafe.Sizeof("")]byte
	lock lockState
	_    [cacheLineSize - unsafe.Sizeof(lockState{})]byte
}

const namedShards = 64

type namedTable [namedShards]atomic.Pointer[namedShard]

// A shard finds its cached lock without the shard lock, so the cached pointer
// has its own cache line, apart from the state that holders of mu write.
type namedShard struct {
	namedShardCache
	_ [(cacheLineSize - unsafe.Sizeof(namedShardCache{})%cacheLineSize) % cacheLineSize]byte
	namedShardData
	_ [(cacheLineSize - unsafe.Sizeof(namedShardData{})%cacheLineSize) % cacheLineSize]byte
}

// cached is an idle-retained lock. It is replaced only while idle, by marking
// it retired under mu. cachedHash is its name hash, truncated to a word and
// kept here so that other names reject it without reading the entry, whose
// cache line its holders write. The two are updated separately; a lookup that sees a stale hash
// falls back to the shard lock, and names are compared.
type namedShardCache struct {
	cachedHash atomic.Uintptr
	cached     atomic.Pointer[namedEntry]
}

// Locks in use, other than the cached one, live in a small table indexed by
// name hash; a map holds any more.
type namedShardData struct {
	mu     lockState // guards everything below and cached replacement
	small  [namedSmall]*namedEntry
	hashes [namedSmall]uint64     // hashes of the names in small
	more   map[string]*namedEntry // locks in use beyond small
	spare  *namedEntry            // a locked entry never published lock-free, for reuse
	misses uint32                 // releases of uncached entries, for cache replacement
	grown  bool                   // more outgrew namedSmallMap; drop it once empty
}

type namedEntryData struct {
	lock lockState
	name string
	hash uint64
	refs int32 // in use: holders and waiters; guarded by the shard lock
	slot int32 // index in small, or -1 when in more
}

type namedEntry struct {
	namedEntryData
	_ [(cacheLineSize - unsafe.Sizeof(namedEntryData{})%cacheLineSize) % cacheLineSize]byte
}

const namedSmall = 4

// After this many releases of uncached locks in a shard, an idle cached lock
// yields its slot, so a name that went cold cannot keep the slot forever.
const namedCacheReplaceInterval = 64

// Maps never shrink, so a shard drops a map that held more entries than this
// once it is empty again. Smaller maps are kept to avoid reallocation.
const namedSmallMap = 8

var namedMutexSeed = maphash.MakeSeed()

// NewNamedMutex creates a named mutex.
func NewNamedMutex() *NamedMutex { return &NamedMutex{} }

// is reports whether f is the lock for name. Checking the suffix first
// rejects long names that share a prefix without comparing every byte.
func (f *namedFirst) is(name string) bool {
	return len(f.name) == len(name) &&
		(len(name) <= 32 || f.name[len(name)-8:] == name[len(name)-8:]) &&
		f.name == name
}

// newShard returns the shard for a name hash, creating it if necessary.
func (nm *NamedMutex) newShard(h uint64) *namedShard {
	table := nm.table.Load()
	if table == nil {
		table = new(namedTable)
		if !nm.table.CompareAndSwap(nil, table) {
			table = nm.table.Load()
		}
	}
	cell := &table[h&(namedShards-1)]
	sh := cell.Load()
	if sh == nil {
		sh = new(namedShard)
		if !cell.CompareAndSwap(nil, sh) {
			sh = cell.Load()
		}
	}
	return sh
}

// existingShard returns the shard for a name hash, or nil.
func (nm *NamedMutex) existingShard(h uint64) *namedShard {
	if table := nm.table.Load(); table != nil {
		return table[h&(namedShards-1)].Load()
	}
	return nil
}

// Lock acquires the named lock, creating it if necessary.
func (nm *NamedMutex) Lock(name string) {
	f := nm.first.Load()
	if f != nil && f.is(name) {
		f.lock.lock()
		return
	}
	nm.acquire(f, name, 0, nil)
}

// TryLock tries to acquire the named lock without waiting.
func (nm *NamedMutex) TryLock(name string) bool {
	f := nm.first.Load()
	if f != nil && f.is(name) {
		return f.lock.tryLock()
	}
	return nm.tryLockSlow(f, name)
}

// TryLockTimeout tries immediately, then waits up to timeout for the named lock.
// A non-positive timeout is equivalent to TryLock.
func (nm *NamedMutex) TryLockTimeout(name string, timeout time.Duration) bool {
	if timeout <= 0 {
		return nm.TryLock(name)
	}
	f := nm.first.Load()
	if f != nil && f.is(name) {
		return f.lock.lockTimeout(timeout)
	}
	return nm.acquire(f, name, timeout, nil)
}

// LockContext acquires the named lock, blocking until it is available or ctx
// is done. It returns nil once the lock is acquired, or ctx.Err() without
// acquiring it. If ctx is already done, LockContext does not try to acquire.
func (nm *NamedMutex) LockContext(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f := nm.first.Load()
	if f != nil && f.is(name) {
		if f.lock.state.CompareAndSwap(0, stateUnit) || f.lock.acquireSlow(1, 0, ctx) == acquired {
			return nil
		}
	} else if nm.acquire(f, name, 0, ctx) {
		return nil
	}
	return ctx.Err()
}

// Unlock releases the named lock. It panics if the name is not locked.
func (nm *NamedMutex) Unlock(name string) {
	f := nm.first.Load()
	if f != nil && f.is(name) {
		if !f.lock.state.CompareAndSwap(stateUnit, 0) && !f.lock.releaseSlow(1) {
			unlockOfUnlocked(name)
		}
		return
	}
	nm.unlockSlow(name)
}

// publishFirst makes a locked lock for name the first name's lock, unless
// another goroutine set the first name; it then returns that lock instead.
func (nm *NamedMutex) publishFirst(name string) (f *namedFirst, locked bool) {
	f = &namedFirst{name: name}
	f.lock.state.Store(stateUnit)
	if nm.first.CompareAndSwap(nil, f) {
		return f, true
	}
	return nm.first.Load(), false
}

// acquire locks name, which is not the name of f, the first-name lock loaded
// by the caller. A positive timeout or a non-nil ctx bounds the wait.
func (nm *NamedMutex) acquire(f *namedFirst, name string, timeout time.Duration, ctx context.Context) bool {
	if f == nil {
		var locked bool
		if f, locked = nm.publishFirst(name); locked {
			return true
		}
		if f.is(name) {
			return f.lock.state.CompareAndSwap(0, stateUnit) || f.lock.acquireSlow(1, timeout, ctx) == acquired
		}
	}
	h := maphash.String(namedMutexSeed, name)
	sh := nm.existingShard(h)
	if sh == nil {
		sh = nm.newShard(h)
	}
	if c := sh.cachedFor(name, h); c != nil && c.lock.state.CompareAndSwap(0, stateUnit) {
		return true
	}
	return sh.acquire(name, h, timeout, ctx)
}

func (nm *NamedMutex) tryLockSlow(f *namedFirst, name string) bool {
	if f == nil {
		var locked bool
		if f, locked = nm.publishFirst(name); locked {
			return true
		}
		if f.is(name) {
			return f.lock.tryLock()
		}
	}
	h := maphash.String(namedMutexSeed, name)
	sh := nm.existingShard(h)
	if sh == nil {
		sh = nm.newShard(h)
	}
	return sh.tryLock(name, h)
}

func (nm *NamedMutex) unlockSlow(name string) {
	h := maphash.String(namedMutexSeed, name)
	if sh := nm.existingShard(h); sh != nil {
		if c := sh.cachedFor(name, h); c != nil && c.lock.state.CompareAndSwap(stateUnit, 0) {
			return
		}
		if sh.release(name, h) {
			return
		}
	}
	unlockOfUnlocked(name)
}

func unlockOfUnlocked(name string) {
	panic("nsync: unlock of unlocked mutex: " + name)
}

func (e *namedEntry) is(name string, h uint64) bool { return e.hash == h && e.name == name }

// cachedFor returns the cached lock if it is the lock for name.
func (sh *namedShard) cachedFor(name string, h uint64) *namedEntry {
	if sh.cachedHash.Load() == uintptr(h) {
		if c := sh.cached.Load(); c != nil && c.is(name, h) {
			return c
		}
	}
	return nil
}

func (sh *namedShard) acquire(name string, h uint64, timeout time.Duration, ctx context.Context) bool {
	for {
		if c := sh.cachedFor(name, h); c != nil {
			if c.lock.state.CompareAndSwap(0, stateUnit) {
				return true
			}
			if r := c.lock.acquireSlow(1, timeout, ctx); r != retired {
				return r == acquired
			}
		}
		sh.mu.lock()
		if c := sh.cached.Load(); c != nil && c.is(name, h) {
			// Cached since the check above; a cached lock is live while mu is held.
			sh.mu.unlock()
			continue
		}
		e := sh.find(name, h)
		if e == nil {
			sh.insert(name, h)
			sh.mu.unlock()
			return true
		}
		e.refs++
		if e.lock.tryLock() {
			sh.mu.unlock()
			return true
		}
		// The reference keeps the entry in use while this caller waits.
		sh.mu.unlock()
		if e.lock.acquireSlow(1, timeout, ctx) == acquired {
			return true
		}
		sh.mu.lock()
		sh.drop(e)
		sh.mu.unlock()
		return false
	}
}

func (sh *namedShard) tryLock(name string, h uint64) bool {
	for {
		if c := sh.cachedFor(name, h); c != nil {
			s := c.lock.state.Load()
			if s&stateRetired == 0 {
				if s > stateQueued {
					return false
				}
				if c.lock.state.CompareAndSwap(s, s+stateUnit) {
					return true
				}
				continue
			}
		}
		sh.mu.lock()
		if c := sh.cached.Load(); c != nil && c.is(name, h) {
			sh.mu.unlock()
			continue
		}
		ok := true
		if e := sh.find(name, h); e == nil {
			sh.insert(name, h)
		} else if ok = e.lock.tryLock(); ok {
			e.refs++
		}
		sh.mu.unlock()
		return ok
	}
}

// release unlocks name, reporting false if it is not locked.
func (sh *namedShard) release(name string, h uint64) bool {
	if c := sh.cachedFor(name, h); c != nil {
		// A held lock cannot be retired, so a retired one is not the caller's.
		if c.lock.state.CompareAndSwap(stateUnit, 0) {
			return true
		}
		if c.lock.state.Load()&stateRetired == 0 {
			return c.lock.releaseSlow(1)
		}
	}
	sh.mu.lock()
	e := sh.find(name, h)
	if e == nil || permits(e.lock.state.Load()) == 0 {
		sh.mu.unlock()
		return false
	}
	if e.refs == 1 {
		// Only the caller refers to the entry, so it has no waiters. A
		// spare stays locked for its next user.
		sh.drop(e)
	} else {
		// Release under mu: a waiter giving up must not see its reference
		// as the last one while the lock is still held.
		e.refs--
		e.lock.unlock()
	}
	sh.mu.unlock()
	return true
}

// find returns the lock in use for name. mu must be held.
func (sh *namedShard) find(name string, h uint64) *namedEntry {
	for i := range sh.small {
		if sh.hashes[i] == h {
			if e := sh.small[i]; e != nil && e.name == name {
				return e
			}
		}
	}
	if sh.more != nil {
		return sh.more[name]
	}
	return nil
}

// insert adds a locked entry for name. mu must be held.
func (sh *namedShard) insert(name string, h uint64) {
	e := sh.spare
	if e == nil {
		e = new(namedEntry)
		e.lock.state.Store(stateUnit)
	} else {
		sh.spare = nil
	}
	e.name, e.hash, e.refs = name, h, 1
	for i := range sh.small {
		if sh.small[i] == nil {
			sh.small[i], sh.hashes[i] = e, h
			e.slot = int32(i)
			return
		}
	}
	e.slot = -1
	if sh.more == nil {
		sh.more = make(map[string]*namedEntry)
	} else if len(sh.more) >= namedSmallMap {
		sh.grown = true
	}
	sh.more[name] = e
}

// drop releases a reference to an entry in use. The last reference removes
// the entry, which is then held by its last user or idle, and known only to
// this shard. mu must be held.
func (sh *namedShard) drop(e *namedEntry) {
	if e.refs--; e.refs != 0 {
		return
	}
	if e.slot >= 0 {
		sh.small[e.slot] = nil
	} else {
		delete(sh.more, e.name)
		if sh.grown && len(sh.more) == 0 {
			sh.more, sh.grown = nil, false
		}
	}
	c := sh.cached.Load()
	if c != nil {
		sh.misses++
	}
	if c == nil || sh.misses%namedCacheReplaceInterval == 0 && c.retire() {
		// A retired entry may still be read by lock-free lookups, so it is
		// never reused.
		e.lock.state.Store(0)
		sh.cachedHash.Store(uintptr(e.hash))
		sh.cached.Store(e)
		return
	}
	if sh.spare == nil {
		e.name = ""
		if e.lock.state.Load() != stateUnit {
			e.lock.state.Store(stateUnit)
		}
		sh.spare = e
	}
}

// retire marks an idle cached entry so that lock-free lookups stop using it.
// It fails while the lock is held or has waiters. The state word alone cannot
// show that: a woken waiter about to retry is still queued. sh.mu must be held.
func (e *namedEntry) retire() bool {
	q := e.lock.queue.Load()
	if q == nil {
		// A goroutine that has yet to queue will see stateRetired.
		return e.lock.state.CompareAndSwap(0, stateRetired)
	}
	q.mu.Lock()
	ok := q.head == nil && e.lock.state.CompareAndSwap(0, stateRetired)
	q.mu.Unlock()
	return ok
}
