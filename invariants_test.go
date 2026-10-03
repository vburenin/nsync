package nsync

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/maphash"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
	"unsafe"
)

func TestLibraryDoesNotUseChannels(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.ChanType, *ast.SendStmt:
				t.Errorf("channel operation in library: %s", fset.Position(node.Pos()))
			case *ast.UnaryExpr:
				if node.Op == token.ARROW {
					t.Errorf("channel receive in library: %s", fset.Position(node.Pos()))
				}
			}
			return true
		})
	}
}

func TestCacheLineLayouts(t *testing.T) {
	for _, size := range []uintptr{unsafe.Sizeof(namedEntry{}), unsafe.Sizeof(namedFirst{}), unsafe.Sizeof(namedShard{}), unsafe.Sizeof(namedOnceShard{})} {
		if size%cacheLineSize != 0 {
			t.Errorf("size %d is not a multiple of cache line size %d", size, cacheLineSize)
		}
	}
	var first namedFirst
	if unsafe.Offsetof(first.lock)%cacheLineSize != 0 {
		t.Error("first lock shares a cache line with its name")
	}
	var shard namedShard
	if unsafe.Offsetof(shard.namedShardData)%cacheLineSize != 0 {
		t.Error("shard state shares a cache line with its cached lock")
	}
	var once NamedOnceMutex
	if unsafe.Offsetof(once.primaryShard)%cacheLineSize != 0 {
		t.Error("primary shard shares a cache line with its index")
	}
}

func TestDisarmWaitsForStartedCallback(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var q waitQueue
		w := getWaiter(&q.mu)
		q.mu.Lock()
		q.pushBack(w)
		started := make(chan struct{})
		timer := time.AfterFunc(0, func() { close(started); w.fire() })
		<-started
		// The callback has started but cannot cancel the waiter while q.mu is
		// held. disarm must release q.mu while waiting for it, then reacquire it.
		w.disarm(waitCancel{timer: timer})
		if q.mu.TryLock() {
			t.Fatal("disarm returned without the queue lock held")
		}
		if !w.canceled || w.queued || q.head != nil {
			t.Fatal("the callback did not remove the waiter")
		}
		q.mu.Unlock()
		putWaiter(w)
	})
}

func TestStarvingMutexHandsOffInOrder(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := NewTryMutex()
		m.Lock()
		entered := make(chan int)
		release := make(chan struct{})
		var workers sync.WaitGroup
		for i := range 8 {
			workers.Go(func() {
				m.Lock()
				entered <- i
				<-release
				m.Unlock()
			})
			synctest.Wait() // queue the waiters in a known order
		}
		// Force handoff mode deterministically, without depending on the host
		// scheduler to exceed the handoff threshold.
		m.state.state.Or(stateStarving)
		m.Unlock()
		if m.TryLock() {
			m.Unlock()
			t.Fatal("a new arrival took a lock handed to a waiter")
		}
		for want := range 8 {
			if got := <-entered; got != want {
				t.Errorf("waiter %d acquired the lock in position %d", got, want)
			}
			release <- struct{}{}
		}
		workers.Wait()
		if s := m.state.state.Load(); s != 0 {
			t.Fatalf("state after all waiters left = %#x, want 0", s)
		}
	})
}

func TestReleaseClearsQueuedFlagOfEmptyQueue(t *testing.T) {
	// A canceled waiter's callback removes it from the queue before the
	// waiter clears stateQueued. A release in between must clear it.
	var m lockState
	m.lock()
	m.waitQueue()
	m.state.Store(stateUnit | stateQueued)
	m.unlock()
	if s := m.state.Load(); s != 0 {
		t.Fatalf("state = %#x, want 0", s)
	}
	if !m.tryLock() {
		t.Fatal("release left the mutex unusable")
	}
	m.unlock()
}

func TestLastCanceledWaiterClearsHandoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var m lockState
		m.lock()
		done := make(chan bool)
		go func() { done <- m.lockTimeout(time.Second) }()
		synctest.Wait()
		m.state.Or(stateStarving)
		if <-done {
			t.Fatal("acquired a held mutex")
		}
		if s := m.state.Load(); s != stateUnit {
			t.Fatalf("state after the last waiter left = %#x, want %#x", s, stateUnit)
		}
		m.unlock()
		if !m.tryLock() {
			t.Fatal("canceled waiter left the mutex unusable")
		}
		m.unlock()
	})
}

func TestNamedMutexCollisionsAndLongKeys(t *testing.T) {
	var m NamedMutex
	const first = "first"
	target := maphash.String(namedMutexSeed, first) & 63
	keys := []string{first}
	for i := 0; len(keys) < 4; i++ {
		key := "collision-" + strconv.Itoa(i)
		if maphash.String(namedMutexSeed, key)&63 == target {
			keys = append(keys, key)
		}
	}
	keys = append(keys, "", strings.Repeat("prefix-", 100)+"one", strings.Repeat("prefix-", 100)+"two")
	for _, key := range keys {
		m.Lock(key)
	}
	for _, key := range keys {
		if m.TryLock(key) {
			t.Errorf("key %q was acquired twice", key)
		}
		m.Unlock(key)
		if !m.TryLockTimeout(key, 0) {
			t.Errorf("key %q could not be reacquired", key)
		}
		m.Unlock(key)
	}
	mustPanic(t, func() { m.Unlock("never-created") })
}

func TestControlWaitGroupMultipleWaiters(t *testing.T) {
	for _, abort := range []bool{false, true} {
		synctest.Test(t, func(t *testing.T) {
			cwg := NewControlWaitGroup(1)
			release := make(chan struct{})
			cwg.Do(func() { <-release })
			var callers sync.WaitGroup
			for range 8 {
				callers.Go(func() { cwg.Do(func() {}) })
			}
			synctest.Wait()
			var waiters sync.WaitGroup
			for range 8 {
				waiters.Go(cwg.Wait)
			}
			synctest.Wait()
			if abort {
				cwg.Abort()
			}
			close(release)
			callers.Wait()
			waiters.Wait()
			if cwg.Working() != 0 || cwg.Waiting() != 0 {
				t.Error("tasks remain after all Wait calls returned")
			}
		})
	}
}

func (nm *NamedMutex) shardOf(name string) *namedShard {
	return nm.existingShard(maphash.String(namedMutexSeed, name))
}

// inUse counts the names locked or awaited in the shard. mu must be held, or
// the shard must be quiescent.
func (sh *namedShard) inUse() int {
	n := len(sh.more)
	for _, e := range sh.small {
		if e != nil {
			n++
		}
	}
	return n
}

// A release without an acquisition wraps the semaphore word negative until
// releaseSlow undoes it and panics. Acquirers that see the negative word must
// wait for the permit rather than take the word for a retired lock.
func TestSemaphoreMisusedReleaseWindow(t *testing.T) {
	acquirers := []struct {
		name    string
		acquire func(*Semaphore) bool
	}{
		{"Acquire", func(s *Semaphore) bool { s.Acquire(); return true }},
		{"AcquireContext", func(s *Semaphore) bool { return s.AcquireContext(context.Background()) == nil }},
		{"TryAcquireTimeout", func(s *Semaphore) bool { return s.TryAcquireTimeout(time.Hour) }},
	}
	for _, a := range acquirers {
		for _, capacity := range []int{1, 2} {
			t.Run(a.name+"/"+strconv.Itoa(capacity), func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					s := NewSemaphore(capacity)
					s.lock.state.Add(^(stateUnit - 1)) // the decrement of the misused Release
					var done, ok atomic.Bool
					go func() {
						ok.Store(a.acquire(s))
						done.Store(true)
					}()
					synctest.Wait()
					if done.Load() {
						t.Fatalf("returned %v while the release was being undone", ok.Load())
					}
					if r := undoMisusedRelease(s); r != "nsync: release without acquisition" {
						t.Errorf("undoing the release panicked with %v", r)
					}
					synctest.Wait()
					if !done.Load() || !ok.Load() {
						t.Fatalf("after the release was undone: returned %v, acquired %v", done.Load(), ok.Load())
					}
					if v := s.Value(); v != 1 {
						t.Fatalf("Value() = %d, want 1", v)
					}
					s.Release()
					if st := s.lock.state.Load(); st != 0 {
						t.Fatalf("state = %#x, want 0", st)
					}
				})
			})
		}
	}
}

// undoMisusedRelease runs the second half of a Release without an acquisition
// and returns its panic value.
func undoMisusedRelease(s *Semaphore) (r any) {
	defer func() { r = recover() }()
	s.releaseSlow()
	return nil
}

// Overlapping misused releases leave the word negative until the last one is
// undone; only then may waiters be woken, and the word must end up idle.
func TestSemaphoreOverlappingMisusedReleases(t *testing.T) {
	for _, contended := range []bool{false, true} {
		t.Run("queue="+strconv.FormatBool(contended), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				s := NewSemaphore(1)
				if contended { // allocate the wait queue, then empty it
					s.Acquire()
					go func() { s.Acquire(); s.Release() }()
					synctest.Wait()
					s.Release()
					synctest.Wait()
				}
				s.lock.state.Add(^(stateUnit - 1))
				s.lock.state.Add(^(stateUnit - 1))
				for range 2 {
					if r := undoMisusedRelease(s); r != "nsync: release without acquisition" {
						t.Fatalf("undoing a release panicked with %v", r)
					}
				}
				if st := s.lock.state.Load(); st != 0 {
					t.Fatalf("state = %#x, want 0", st)
				}
				if !s.TryAcquire() {
					t.Fatal("TryAcquire failed on an idle semaphore")
				}
				s.Release()
			})
		})
	}
}

// An acquirer that gives up while a misused release is being undone must not
// clear bits of the wrapped count.
func TestSemaphoreWaiterGivesUpDuringMisusedRelease(t *testing.T) {
	acquirers := []struct {
		name    string
		acquire func(*Semaphore) bool
	}{
		{"TryAcquireTimeout", func(s *Semaphore) bool { return s.TryAcquireTimeout(time.Second) }},
		{"AcquireContext", func(s *Semaphore) bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			return s.AcquireContext(ctx) == nil
		}},
	}
	for _, a := range acquirers {
		t.Run(a.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				s := NewSemaphore(1)
				s.lock.state.Add(^(stateUnit - 1))
				var done, ok atomic.Bool
				go func() {
					ok.Store(a.acquire(s))
					done.Store(true)
				}()
				time.Sleep(2 * time.Second)
				synctest.Wait()
				if !done.Load() || ok.Load() {
					t.Fatalf("returned %v, acquired %v; want a timeout", done.Load(), ok.Load())
				}
				if r := undoMisusedRelease(s); r != "nsync: release without acquisition" {
					t.Fatalf("undoing the release panicked with %v", r)
				}
				if st := s.lock.state.Load(); st != 0 {
					t.Fatalf("state = %#x, want 0", st)
				}
			})
		})
	}
}

// On 32-bit platforms a shard keeps only the low 32 bits of its cached lock's
// name hash. Names whose hashes share those bits must still get their own locks.
func TestNamedMutexPartialHashCollision(t *testing.T) {
	seen := make(map[uint32]string)
	var a, b string
	for i := 0; a == ""; i++ {
		name := "n" + strconv.Itoa(i)
		h := uint32(maphash.String(namedMutexSeed, name))
		if other, ok := seen[h]; ok {
			a, b = other, name
		}
		seen[h] = name
	}
	var m NamedMutex
	m.Lock("first") // the first name has its own lock; keep it out of the way
	m.Unlock("first")
	m.Lock(a)
	m.Unlock(a) // a's lock becomes its shard's cached lock
	if sh := m.shardOf(a); sh == nil || sh.cached.Load() == nil || sh.cached.Load().name != a {
		t.Fatalf("%q is not the cached lock of its shard", a)
	}
	m.Lock(a)
	if !m.TryLock(b) {
		t.Fatalf("%q and %q share a lock", a, b)
	}
	if m.TryLock(a) || m.TryLock(b) {
		t.Fatal("a held name was locked again")
	}
	m.Unlock(b)
	// While a shard replaces its cached lock, a lock-free lookup can see the
	// new hash beside the old lock. Another name must not use that lock either.
	sh := m.shardOf(a)
	sh.cachedHash.Store(uintptr(maphash.String(namedMutexSeed, b)))
	if !m.TryLock(b) {
		t.Fatalf("%q used %q's lock through a stale cached hash", b, a)
	}
	m.Unlock(b)
	sh.cachedHash.Store(uintptr(maphash.String(namedMutexSeed, a)))
	m.Unlock(a)
	checkNamedMutexIdle(t, &m)
}
