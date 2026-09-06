package nsync

import (
	"go/ast"
	"go/parser"
	"go/token"
	"hash/maphash"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	for _, size := range []uintptr{unsafe.Sizeof(namedMutexEntry{}), unsafe.Sizeof(namedPrimaryEntry{}), unsafe.Sizeof(namedOnceShard{})} {
		if size%cacheLineSize != 0 {
			t.Errorf("size %d is not a multiple of cache line size %d", size, cacheLineSize)
		}
	}
	var primary namedPrimaryEntry
	if unsafe.Offsetof(primary.mutex)%cacheLineSize != 0 {
		t.Error("primary mutex shares a cache line with its name")
	}
	var once NamedOnceMutex
	if unsafe.Offsetof(once.primaryShard)%cacheLineSize != 0 {
		t.Error("primary shard shares a cache line with its index")
	}
}

func TestStopWaitsForStartedTimerCallback(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var mu sync.Mutex
		cond := sync.NewCond(&mu)
		w := &timedWait{cond: cond}
		w.callback = w.expire
		mu.Lock()
		started := make(chan struct{})
		timer := time.AfterFunc(0, func() { close(started); w.expire() })
		<-started
		// The callback has started but cannot set expired while mu is held.
		// stop must release mu while waiting for it, then reacquire mu.
		w.stop(timer)
		if mu.TryLock() {
			t.Fatal("stop returned without the condition mutex held")
		}
		mu.Unlock()
	})
}

func TestMutexHandoffRejectsBarging(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := NewTryMutex()
		m.Lock()
		entered := make(chan struct{})
		release := make(chan struct{})
		var workers sync.WaitGroup
		for range 8 {
			workers.Go(func() {
				m.Lock()
				entered <- struct{}{}
				<-release
				m.Unlock()
			})
		}
		synctest.Wait()
		// Force the slow-path condition deterministically, without depending
		// on the host scheduler to exceed the handoff threshold.
		m.state.mu.Lock()
		m.state.fair = true
		m.state.mu.Unlock()
		m.Unlock()
		if m.TryLock() {
			m.Unlock()
			t.Fatal("a new arrival stole a reserved handoff")
		}
		for range 8 {
			<-entered
			release <- struct{}{}
		}
		workers.Wait()
		if !m.TryLock() {
			t.Fatal("handoff did not return the mutex to an idle state")
		}
		m.Unlock()
	})
}

func TestLastTimedWaiterMayLeaveBeforeUnlockSlow(t *testing.T) {
	// unlock's CAS can fail on state 2, then the last timed waiter can leave
	// and restore state 1 before unlock obtains the queue mutex.
	var m mutexState
	m.lock()
	m.unlockSlow()
	if !m.tryLock() {
		t.Fatal("release after the last waiter left retained the lock")
	}
	m.unlock()
}

func TestLastCanceledWaiterClearsHandoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Exercise a reserved lock with no owner. A new timed arrival cannot
		// take the reservation; leaving as the last waiter must release it.
		var m mutexState
		m.state.Store(3)
		if m.lockTimeout(time.Nanosecond) {
			t.Fatal("new arrival consumed a reserved handoff")
		}
		if !m.tryLock() {
			t.Fatal("canceled waiter left an ownerless reservation")
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
