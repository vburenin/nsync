package nsync

import (
	"fmt"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected a panic")
		}
	}()
	f()
}

func TestTimeoutImmediateAcquisition(t *testing.T) {
	for _, timeout := range []time.Duration{-time.Second, 0, time.Nanosecond} {
		t.Run(timeout.String(), func(t *testing.T) {
			mutex := NewTryMutex()
			semaphore := NewSemaphore(1)
			named := NewNamedMutex()
			for _, lock := range []struct {
				name    string
				acquire func(time.Duration) bool
				release func()
			}{
				{"TryMutex", mutex.TryLockTimeout, mutex.Unlock},
				{"Semaphore", semaphore.TryAcquireTimeout, semaphore.Release},
				{"NamedMutex", func(d time.Duration) bool { return named.TryLockTimeout("key", d) }, func() { named.Unlock("key") }},
			} {
				t.Run(lock.name, func(t *testing.T) {
					for range 100 {
						if !lock.acquire(timeout) {
							t.Fatal("an available lock must be acquired immediately")
						}
						if lock.acquire(timeout) {
							t.Fatal("acquired an already held lock")
						}
						lock.release()
					}
				})
			}
		})
	}
}

func TestAbortUnblocksPending(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cwg := NewControlWaitGroup(1)
		release := make(chan struct{})
		defer cwg.Wait()
		defer close(release)
		cwg.Do(func() { <-release })

		const pending = 8
		results := make(chan bool, pending)
		for range pending {
			go func() { results <- cwg.Do(func() { t.Error("aborted task ran") }) }()
		}
		synctest.Wait()
		if got := cwg.Waiting(); got != pending {
			t.Fatalf("Waiting() = %d, want %d", got, pending)
		}
		cwg.Abort()
		cwg.Abort() // Aborting repeatedly must be safe.
		synctest.Wait()
		for range pending {
			select {
			case ran := <-results:
				if ran {
					t.Error("Do returned true for an aborted task")
				}
			default:
				t.Fatal("Abort did not unblock a pending Do while a worker was still running")
			}
		}
		if got := cwg.Waiting(); got != 0 {
			t.Errorf("Waiting() = %d after Abort, want 0", got)
		}
		if got := cwg.Working(); got != 1 {
			t.Errorf("Working() = %d after Abort, want 1", got)
		}
	})
}

func TestDoAfterAbortReturnsImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cwg := NewControlWaitGroup(1)
		release := make(chan struct{})
		defer cwg.Wait()
		defer close(release)
		cwg.Do(func() { <-release })
		cwg.Abort()
		result := make(chan bool, 1)
		go func() { result <- cwg.Do(func() { t.Error("task ran after Abort") }) }()
		synctest.Wait()
		select {
		case ran := <-result:
			if ran {
				t.Error("Do returned true after Abort")
			}
		default:
			t.Fatal("Do blocked after Abort")
		}
	})
}

func TestPositiveCapacityRequired(t *testing.T) {
	for _, capacity := range []int{0, -1} {
		t.Run(fmt.Sprint(capacity), func(t *testing.T) {
			mustPanic(t, func() { NewSemaphore(capacity) })
			mustPanic(t, func() { NewControlWaitGroup(capacity) })
		})
	}
}

func TestNamedOnceMutexInvalidKeyRecovery(t *testing.T) {
	for _, operation := range []string{"Lock", "Unlock"} {
		t.Run(operation, func(t *testing.T) {
			nm := NewNamedOnceMutex()
			mustPanic(t, func() {
				if operation == "Lock" {
					nm.Lock([]int{1})
				} else {
					nm.Unlock([]int{1})
				}
			})
			if !nm.Lock("valid") {
				t.Fatal("valid lock failed after recovering from invalid key")
			}
			nm.Unlock("valid")
		})
	}
}

func TestNamedMutexZeroValue(t *testing.T) {
	var nm NamedMutex
	nm.Lock("lock")
	nm.Unlock("lock")
	if !nm.TryLock("try") {
		t.Fatal("zero-value NamedMutex could not acquire a new name")
	}
	nm.Unlock("try")
	if !nm.TryLockTimeout("timeout", 0) {
		t.Fatal("zero-value NamedMutex could not acquire a new name with timeout")
	}
	nm.Unlock("timeout")
}

func TestNamedOnceMutexZeroValue(t *testing.T) {
	var nm NamedOnceMutex
	for range 2 {
		if !nm.Lock(nil) {
			t.Fatal("could not acquire a nil key")
		}
		if !nm.Lock("other") {
			t.Fatal("a different key must be independent")
		}
		nm.Unlock(nil)
		nm.Unlock("other")
	}
	nm.Unlock("missing")
	shards := []*namedOnceShard{&nm.primaryShard}
	if table := nm.table.Load(); table != nil {
		for i := range table {
			if sh := table[i].Load(); sh != nil {
				shards = append(shards, sh)
			}
		}
	}
	for _, sh := range shards {
		if sh.primary != nil || sh.primaryKey != nil || len(sh.locks) != 0 {
			t.Error("unlocked mutexes were not discarded")
		}
	}
}

func TestNamedMutexConcurrent(t *testing.T) {
	var nm NamedMutex
	var wg sync.WaitGroup
	var count int
	for range 16 {
		wg.Go(func() {
			for range 100 {
				nm.Lock("shared")
				count++
				nm.Unlock("shared")
			}
		})
	}
	wg.Wait()
	if count != 1600 {
		t.Errorf("count = %d, want 1600", count)
	}
}

func TestReleaseWithoutAcquisition(t *testing.T) {
	mustPanic(t, NewTryMutex().Unlock)
	mustPanic(t, NewSemaphore(1).Release)
	var nm NamedMutex
	mustPanic(t, func() { nm.Unlock("missing") })
	// Recovering from the panic must leave the map mutex usable.
	nm.Lock("key")
	nm.Unlock("key")
	mustPanic(t, func() { nm.Unlock("key") })
}

func TestTimedAcquisitionWaitsForRelease(t *testing.T) {
	for _, name := range []string{"TryMutex", "Semaphore", "NamedMutex"} {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var acquire func(time.Duration) bool
				var release func()
				switch name {
				case "TryMutex":
					m := NewTryMutex()
					acquire, release = m.TryLockTimeout, m.Unlock
				case "Semaphore":
					s := NewSemaphore(1)
					acquire, release = s.TryAcquireTimeout, s.Release
				case "NamedMutex":
					m := NewNamedMutex()
					acquire = func(d time.Duration) bool { return m.TryLockTimeout("key", d) }
					release = func() { m.Unlock("key") }
				}
				if !acquire(0) {
					t.Fatal("initial acquisition failed")
				}
				go func() {
					time.Sleep(time.Second)
					release()
				}()
				start := time.Now()
				if !acquire(2 * time.Second) {
					t.Fatal("did not acquire after release")
				}
				if elapsed := time.Since(start); elapsed != time.Second {
					t.Errorf("waited %v, want 1s", elapsed)
				}
				start = time.Now()
				if acquire(time.Second) {
					t.Fatal("acquired a held lock")
				}
				if elapsed := time.Since(start); elapsed != time.Second {
					t.Errorf("timed out after %v, want 1s", elapsed)
				}
				release()
			})
		})
	}
}
