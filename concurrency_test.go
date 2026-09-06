package nsync

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestTryMutexValueCopy(t *testing.T) {
	m := NewTryMutex()
	copy := *m
	var _ sync.Locker = copy
	m.Lock()
	if copy.TryLock() {
		t.Fatal("copy did not share the lock")
	}
	copy.Unlock()
	if !m.TryLock() {
		t.Fatal("unlock through copy was not visible")
	}
	m.Unlock()
}

func TestTryMutexMutualExclusionAndPublication(t *testing.T) {
	m := NewTryMutex()
	var count, checksum int
	var wg sync.WaitGroup
	for range 32 {
		wg.Go(func() {
			for range 200 {
				m.Lock()
				if checksum != count*17 {
					t.Error("protected writes were not published")
				}
				count++
				runtime.Gosched()
				checksum = count * 17
				m.Unlock()
			}
		})
	}
	wg.Wait()
	if count != 6400 {
		t.Fatalf("count = %d, want 6400", count)
	}
}

func TestSemaphoreConcurrentCapacity(t *testing.T) {
	for _, capacity := range []int{1, 2, 8, 64} {
		s := NewSemaphore(capacity)
		var active atomic.Int32
		var wg sync.WaitGroup
		for range 64 {
			wg.Go(func() {
				for range 50 {
					s.Acquire()
					if n := active.Add(1); n > int32(capacity) {
						t.Errorf("active = %d, capacity = %d", n, capacity)
					}
					runtime.Gosched()
					active.Add(-1)
					s.Release()
				}
			})
		}
		wg.Wait()
		if s.Value() != 0 || active.Load() != 0 {
			t.Fatal("semaphore leaked an acquisition")
		}
	}
}

func TestSemaphoreLargeCapacity(t *testing.T) {
	s := NewSemaphore(math.MaxInt)
	s.Acquire()
	if !s.TryAcquire() || s.Value() != 2 {
		t.Fatal("large capacity does not work")
	}
	s.Release()
	s.Release()
}

func TestGeneralSemaphoreNonblockingAndInvalidRelease(t *testing.T) {
	s := NewSemaphore(8)
	mustPanic(t, s.Release)
	for range 8 {
		if !s.TryAcquireTimeout(0) {
			t.Fatal("nonblocking acquisition rejected an available slot")
		}
	}
	for _, timeout := range []time.Duration{0, -time.Second} {
		if s.TryAcquireTimeout(timeout) {
			t.Fatal("nonblocking acquisition exceeded capacity")
		}
	}
	for range 8 {
		s.Release()
	}
	mustPanic(t, s.Release)
	if !s.TryAcquire() {
		t.Fatal("invalid release corrupted the counter")
	}
	s.Release()
}

func TestNamedMutexConcurrentCreation(t *testing.T) {
	for range 8 {
		var m NamedMutex
		var keys [16]string
		var counts [16]int
		for i := range keys {
			keys[i] = fmt.Sprintf("%s%d", strings.Repeat("prefix", i), i)
		}
		start := make(chan struct{})
		var workers sync.WaitGroup
		for i := range 64 {
			workers.Go(func() {
				<-start
				index := i % len(keys)
				for range 20 {
					m.Lock(keys[index])
					counts[index]++
					runtime.Gosched()
					m.Unlock(keys[index])
				}
			})
		}
		close(start)
		workers.Wait()
		for _, count := range counts {
			if count != 80 {
				t.Fatalf("concurrent creation produced independent locks for one name: count %d", count)
			}
		}
	}
}

func TestNamedOnceConcurrentKeys(t *testing.T) {
	var m NamedOnceMutex
	var active [16]atomic.Int32
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := range 64 {
		workers.Go(func() {
			<-start
			key := i % len(active)
			for range 50 {
				if m.Lock(key) {
					if active[key].Add(1) != 1 {
						t.Error("concurrent key creation admitted duplicate operations")
					}
					runtime.Gosched()
					active[key].Add(-1)
					m.Unlock(key)
				}
			}
		})
	}
	close(start)
	workers.Wait()
}

func TestTimedWaitersDoNotStrandBlockingWaiter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := NewTryMutex()
		m.Lock()
		var timed sync.WaitGroup
		for range 16 {
			timed.Go(func() {
				if m.TryLockTimeout(time.Second) {
					m.Unlock()
					t.Error("acquired before release")
				}
			})
		}
		finished := make(chan struct{})
		go func() { m.Lock(); m.Unlock(); close(finished) }()
		synctest.Wait()
		time.Sleep(time.Second)
		timed.Wait()
		m.Unlock()
		<-finished
		if !m.TryLock() {
			t.Fatal("a canceled waiter retained the lock")
		}
		m.Unlock()
	})
}

func TestTimedWaitersDifferentDeadlines(t *testing.T) {
	for _, capacity := range []int{1, 2, 8} {
		synctest.Test(t, func(t *testing.T) {
			s := NewSemaphore(capacity)
			for range capacity {
				s.Acquire()
			}
			var acquired, expired atomic.Int32
			var wg sync.WaitGroup
			for i := 1; i <= 8; i++ {
				wg.Go(func() {
					if s.TryAcquireTimeout(time.Duration(i) * time.Millisecond) {
						acquired.Add(1)
						s.Release()
					} else {
						expired.Add(1)
					}
				})
			}
			synctest.Wait()
			time.Sleep(4500 * time.Microsecond)
			for range capacity {
				s.Release()
			}
			wg.Wait()
			if acquired.Load() != 4 || expired.Load() != 4 {
				t.Errorf("acquired = %d, expired = %d, want 4 each", acquired.Load(), expired.Load())
			}
			if s.Value() != 0 {
				t.Fatal("timed wait leaked a slot")
			}
		})
	}

}

func TestTimeoutConcurrentWithRelease(t *testing.T) {
	for _, capacity := range []int{1, 2, 8} {
		for range 200 {
			s := NewSemaphore(capacity)
			for range capacity {
				s.Acquire()
			}
			var wg sync.WaitGroup
			wg.Go(func() {
				if s.TryAcquireTimeout(time.Microsecond) {
					s.Release()
				}
			})
			wg.Go(func() {
				for range capacity {
					s.Release()
				}
			})
			wg.Wait()
			if s.Value() != 0 {
				t.Fatal("release/timeout race leaked a slot")
			}
			if !s.TryAcquire() {
				t.Fatal("semaphore unusable after release/timeout race")
			}
			s.Release()
		}
	}

}

func TestOnceMutexPublishesToEveryWaiter(t *testing.T) {
	var m OnceMutex
	m.Lock()
	var value int
	var wg sync.WaitGroup
	for range 64 {
		wg.Go(func() {
			if m.Lock() {
				t.Error("OnceMutex acquired twice")
				m.Unlock()
				return
			}
			if value != 42 {
				t.Error("OnceMutex did not publish protected data")
			}
		})
	}
	value = 42
	m.Unlock()
	wg.Wait()
}

func TestNamedOnceMutexConcurrentExclusion(t *testing.T) {
	var m NamedOnceMutex
	var wg sync.WaitGroup
	var active atomic.Int32
	for range 32 {
		wg.Go(func() {
			for range 100 {
				if m.Lock("shared") {
					if active.Add(1) != 1 {
						t.Error("overlapping successful named-once acquisitions")
					}
					runtime.Gosched()
					active.Add(-1)
					m.Unlock("shared")
				}
			}
		})
	}
	wg.Wait()
}

func TestNamedOnceMutexGenerationIsolation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var m NamedOnceMutex
		if !m.Lock("key") {
			t.Fatal("first operation was not admitted")
		}
		results := make(chan bool, 64)
		for range 64 {
			go func() { results <- m.Lock("key") }()
		}
		synctest.Wait()
		m.Unlock("key")
		if !m.Lock("key") {
			t.Fatal("new operation was not admitted")
		}
		synctest.Wait()
		for range 64 {
			if <-results {
				t.Error("old waiter acquired the new operation")
			}
		}
		m.Unlock("key")
		for range 100 {
			if !m.Lock("key") {
				t.Fatal("reused entry was still completed")
			}
			m.Unlock("key")
		}
	})
}

func TestControlWaitGroupGoexit(t *testing.T) {
	cwg := NewControlWaitGroup(1)
	cwg.Do(runtime.Goexit)
	cwg.Wait()
	if !cwg.Do(func() {}) {
		t.Fatal("group unusable after task exits")
	}
	cwg.Wait()
	if cwg.Working() != 0 || cwg.Waiting() != 0 {
		t.Fatal("Goexit leaked a task or slot")
	}
}

func TestOnceMutexInvalidUnlock(t *testing.T) {
	var m OnceMutex
	mustPanic(t, m.Unlock)
	if !m.Lock() {
		t.Fatal("invalid unlock changed the unused state")
	}
	m.Unlock()
	mustPanic(t, m.Unlock)
	if m.Lock() {
		t.Fatal("invalid unlock reset the completed state")
	}
}

func TestTinyTimeoutsUseTheClock(t *testing.T) {
	for _, timeout := range []time.Duration{time.Nanosecond, 100 * time.Nanosecond, time.Microsecond} {
		synctest.Test(t, func(t *testing.T) {
			m := NewTryMutex()
			m.Lock()
			start := time.Now()
			if m.TryLockTimeout(timeout) {
				t.Fatal("acquired a held mutex")
			}
			if elapsed := time.Since(start); elapsed != timeout {
				t.Errorf("waited %v, want %v", elapsed, timeout)
			}
			m.Unlock()
			s := NewSemaphore(2)
			s.Acquire()
			s.Acquire()
			start = time.Now()
			if s.TryAcquireTimeout(timeout) {
				t.Fatal("acquired a full semaphore")
			}
			if elapsed := time.Since(start); elapsed != timeout {
				t.Errorf("waited %v, want %v", elapsed, timeout)
			}
			s.Release()
			s.Release()
		})
	}
}

func TestNamedOnceMutexCollidingKeys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var m NamedOnceMutex
		keys := []int{0}
		sh := m.shard(0, true)
		for key := 1; len(keys) < 3; key++ {
			if m.shard(key, true) == sh {
				keys = append(keys, key)
			}
		}
		if !m.Lock(keys[0]) || !m.Lock(keys[1]) {
			t.Fatal("distinct colliding keys must be independent")
		}
		results := make(chan bool, 2)
		go func() { results <- m.Lock(keys[1]) }()
		synctest.Wait()
		m.Unlock(keys[0])
		// The second key remains active even though the primary slot is free.
		go func() { results <- m.Lock(keys[1]) }()
		synctest.Wait()
		if !m.Lock(keys[2]) {
			t.Fatal("third colliding key was not admitted")
		}
		m.Unlock(keys[1])
		for range 2 {
			if <-results {
				t.Error("a waiter started a duplicate operation")
			}
		}
		m.Unlock(keys[2])
		if !m.Lock(keys[1]) {
			t.Fatal("completed key could not be reused")
		}
		m.Unlock(keys[1])
		if sh.primary != nil || sh.primaryKey != nil || sh.extraCount != 0 || len(sh.locks) != 0 {
			t.Error("completed keys were retained")
		}
	})
}
