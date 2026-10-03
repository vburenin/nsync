package nsync

import (
	"context"
	"strconv"
	"testing"
	"testing/synctest"
	"time"
)

func TestLock(t *testing.T) {
	l := NewNamedMutex()
	l.Lock("test1")
	l.Lock("test2")
	if l.TryLock("test1") || l.TryLock("test2") {
		t.Error("Lock acquired twice")
	}
	l.Unlock("test1")
	l.Unlock("test2")

	if !l.TryLock("test1") {
		t.Error("Lock is not released")
	}
	if !l.TryLock("test2") {
		t.Error("Lock is not released")
	}
	l.Unlock("test1")
	l.Unlock("test2")
}

func TestLock2(t *testing.T) {
	l := NewNamedMutex()
	l.Lock("test1")
	go func() {
		time.Sleep(time.Millisecond * 10)
		l.Unlock("test1")
	}()
	l.Lock("test1")
	l.Unlock("test1")
}

func TestTryLock(t *testing.T) {
	l := NewNamedMutex()
	if !l.TryLock("test1") {
		t.Error("Didn't acquire lock")
	}
	if !l.TryLock("test2") {
		t.Error("Didn't acquire lock")
	}

	if l.TryLock("test1") {
		t.Error("Lock should be acquired")
	}
	if l.TryLock("test2") {
		t.Error("Lock should be acquired")
	}
}

func TestTryLockTimeout(t *testing.T) {
	l := NewNamedMutex()
	if !l.TryLockTimeout("test1", time.Millisecond) {
		t.Error("Didn't acquire lock")
	}
	if !l.TryLockTimeout("test2", time.Millisecond) {
		t.Error("Didn't acquire lock")
	}

	if l.TryLockTimeout("test1", time.Millisecond) {
		t.Error("Lock should be acquired")
	}
	if l.TryLockTimeout("test2", time.Millisecond) {
		t.Error("Lock should be acquired")
	}
}

// retainedNamedLocks counts the locks a NamedMutex keeps: names in use plus
// idle cached and spare locks.
func retainedNamedLocks(m *NamedMutex) (inUse, idle int) {
	if m.first.Load() != nil {
		idle++
	}
	table := m.table.Load()
	if table == nil {
		return inUse, idle
	}
	for i := range table {
		sh := table[i].Load()
		if sh == nil {
			continue
		}
		sh.mu.lock()
		inUse += sh.inUse()
		if sh.cached.Load() != nil {
			idle++
		}
		if sh.spare != nil {
			idle++
		}
		sh.mu.unlock()
	}
	return inUse, idle
}

func TestNamedMutexDiscardsUnlockedNames(t *testing.T) {
	var m NamedMutex
	for i := range 100_000 {
		name := "user-" + strconv.Itoa(i)
		switch i % 3 {
		case 0:
			m.Lock(name)
		case 1:
			if !m.TryLock(name) {
				t.Fatalf("TryLock(%q) failed", name)
			}
		default:
			if !m.TryLockTimeout(name, time.Second) {
				t.Fatalf("TryLockTimeout(%q) failed", name)
			}
		}
		m.Unlock(name)
	}
	inUse, idle := retainedNamedLocks(&m)
	if inUse != 0 {
		t.Errorf("%d unlocked names are still in use", inUse)
	}
	if limit := 1 + 2*namedShards; idle > limit {
		t.Errorf("%d idle locks retained, want at most %d", idle, limit)
	}
	mustPanic(t, func() { m.Unlock("user-5") })
}

func TestNamedMutexDiscardsNamesAfterConcurrentUse(t *testing.T) {
	var m NamedMutex
	const names = 10_000
	for i := range names {
		m.Lock(strconv.Itoa(i))
	}
	if inUse, _ := retainedNamedLocks(&m); inUse < names-1-namedShards {
		t.Fatalf("only %d of %d held names are in use", inUse, names)
	}
	for i := range names {
		m.Unlock(strconv.Itoa(i))
	}
	inUse, idle := retainedNamedLocks(&m)
	if inUse != 0 || idle > 1+2*namedShards {
		t.Fatalf("after unlocking: %d in use, %d idle", inUse, idle)
	}
	// Maps that grew are dropped once empty, since Go maps never shrink.
	table := m.table.Load()
	for i := range table {
		if sh := table[i].Load(); sh != nil && sh.more != nil && sh.grown {
			t.Errorf("shard %d keeps a grown map after all its names were unlocked", i)
		}
	}
}

func TestNamedMutexAbandonedWaitersDiscardName(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var m NamedMutex
		m.Lock("first")
		m.Lock("key")
		ctx, cancel := context.WithCancel(context.Background())
		results := make(chan bool, 3)
		go func() { results <- m.TryLockTimeout("key", time.Second) }()
		go func() { results <- m.LockContext(ctx, "key") == nil }()
		go func() { results <- m.TryLock("key") }()
		synctest.Wait()
		cancel()
		time.Sleep(time.Second)
		for range 3 {
			if <-results {
				t.Fatal("acquired a held name")
			}
		}
		m.Unlock("key")
		m.Unlock("first")
		if inUse, _ := retainedNamedLocks(&m); inUse != 0 {
			t.Fatalf("%d names in use after their waiters gave up", inUse)
		}
		if !m.TryLock("key") {
			t.Fatal("the name could not be locked again")
		}
		m.Unlock("key")
	})
}

func TestNamedMutexCachedLockWithWaitersIsNotRetired(t *testing.T) {
	// A woken waiter leaves the lock word idle until it retries, so retiring
	// a cached lock must also check its queue.
	synctest.Test(t, func(t *testing.T) {
		var m NamedMutex
		m.Lock("first")
		m.Unlock("first")
		m.Lock("key")
		m.Unlock("key") // "key" becomes its shard's cached lock
		sh := m.shardOf("key")
		c := sh.cached.Load()
		if c == nil || c.name != "key" {
			t.Fatal("key was not cached")
		}
		m.Lock("key")
		acquired := make(chan struct{})
		go func() {
			m.Lock("key")
			close(acquired)
		}()
		synctest.Wait()
		sh.mu.lock()
		retired := c.retire()
		sh.mu.unlock()
		if retired {
			t.Fatal("retired a cached lock with a queued waiter")
		}
		// The release wakes the waiter and leaves the word idle until the
		// waiter runs; it has either retried already or is still queued.
		m.Unlock("key")
		sh.mu.lock()
		retired = c.retire()
		sh.mu.unlock()
		if retired {
			t.Fatal("retired a cached lock with a woken waiter")
		}
		<-acquired
		m.Unlock("key")
	})
}
