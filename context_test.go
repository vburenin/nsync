package nsync

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

// contextLock adapts each blocking primitive to a common shape.
type contextLock struct {
	name        string
	acquire     func(context.Context) error
	try         func() bool
	release     func()
	stateIsIdle func() bool
}

func contextLocks() []contextLock {
	m := NewTryMutex()
	s := NewSemaphore(1)
	s2 := NewSemaphore(2)
	var nm NamedMutex
	nm.Lock("first") // make "key" use a shard rather than the first-name slot
	nm.Unlock("first")
	return []contextLock{
		{"TryMutex", m.LockContext, m.TryLock, m.Unlock, func() bool { return m.state.state.Load() == 0 }},
		{"Semaphore", s.AcquireContext, s.TryAcquire, s.Release, func() bool { return s.lock.state.Load() == 0 }},
		{"Semaphore2", func(ctx context.Context) error {
			if err := s2.AcquireContext(ctx); err != nil {
				return err
			}
			// Hold one slot permanently, so the second acquisition must wait.
			return nil
		}, s2.TryAcquire, s2.Release, func() bool { return s2.lock.state.Load() == 0 && s2.gate.state.Load() == 0 }},
		{"NamedMutex", func(ctx context.Context) error { return nm.LockContext(ctx, "key") },
			func() bool { return nm.TryLock("key") }, func() { nm.Unlock("key") },
			func() bool {
				sh := nm.shardOf("key")
				return sh != nil && sh.inUse() == 0 && sh.mu.state.Load() == 0
			}},
	}
}

func TestLockContextAcquiresFreeLock(t *testing.T) {
	for _, l := range contextLocks() {
		t.Run(l.name, func(t *testing.T) {
			if err := l.acquire(context.Background()); err != nil {
				t.Fatalf("acquire: %v", err)
			}
			l.release()
			if !l.stateIsIdle() {
				t.Fatal("release did not restore the idle state")
			}
		})
	}
}

func TestLockContextDoneContextDoesNotAcquire(t *testing.T) {
	for _, l := range contextLocks() {
		t.Run(l.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := l.acquire(ctx); !errors.Is(err, context.Canceled) {
				t.Fatalf("acquire with a canceled context = %v, want context.Canceled", err)
			}
			if !l.try() {
				t.Fatal("a failed acquisition left the lock held")
			}
			l.release()
		})
	}
}

func TestLockContextCancelWhileWaiting(t *testing.T) {
	for _, l := range contextLocks() {
		t.Run(l.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				// Fill the lock, including the second slot of Semaphore2.
				var held int
				for l.try() {
					held++
				}
				ctx, cancel := context.WithCancel(context.Background())
				result := make(chan error, 1)
				go func() { result <- l.acquire(ctx) }()
				synctest.Wait()
				select {
				case err := <-result:
					t.Fatalf("acquire returned %v while the lock was held", err)
				default:
				}
				cancel()
				if err := <-result; !errors.Is(err, context.Canceled) {
					t.Fatalf("acquire = %v, want context.Canceled", err)
				}
				for range held {
					l.release()
				}
				if !l.stateIsIdle() {
					t.Fatal("a canceled waiter left the lock state dirty")
				}
			})
		})
	}
}

func TestLockContextDeadlineWhileWaiting(t *testing.T) {
	for _, l := range contextLocks() {
		t.Run(l.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var held int
				for l.try() {
					held++
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				start := time.Now()
				if err := l.acquire(ctx); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("acquire = %v, want context.DeadlineExceeded", err)
				}
				if elapsed := time.Since(start); elapsed != time.Second {
					t.Errorf("waited %v, want 1s", elapsed)
				}
				for range held {
					l.release()
				}
				if !l.stateIsIdle() {
					t.Fatal("an expired waiter left the lock state dirty")
				}
			})
		})
	}
}

func TestLockContextAcquiresAfterRelease(t *testing.T) {
	for _, l := range contextLocks() {
		t.Run(l.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var held int
				for l.try() {
					held++
				}
				go func() {
					time.Sleep(time.Second)
					l.release()
				}()
				ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
				defer cancel()
				start := time.Now()
				if err := l.acquire(ctx); err != nil {
					t.Fatalf("acquire = %v, want nil", err)
				}
				if elapsed := time.Since(start); elapsed != time.Second {
					t.Errorf("waited %v, want 1s", elapsed)
				}
				for range held {
					l.release()
				}
				if !l.stateIsIdle() {
					t.Fatal("state is dirty after all releases")
				}
			})
		})
	}
}

func TestLockContextCancellationStorm(t *testing.T) {
	// Waiters sharing a context are all canceled at once. Each must leave
	// without disturbing the others, and the lock must remain usable.
	synctest.Test(t, func(t *testing.T) {
		m := NewTryMutex()
		m.Lock()
		ctx, cancel := context.WithCancel(context.Background())
		results := make(chan error, 100)
		for range 100 {
			go func() { results <- m.LockContext(ctx) }()
		}
		survivor := make(chan error, 1)
		go func() { survivor <- m.LockContext(context.Background()) }()
		synctest.Wait()
		cancel()
		for range 100 {
			if err := <-results; !errors.Is(err, context.Canceled) {
				t.Fatalf("LockContext = %v, want context.Canceled", err)
			}
		}
		m.Unlock()
		if err := <-survivor; err != nil {
			t.Fatalf("uncanceled waiter: %v", err)
		}
		m.Unlock()
		if s := m.state.state.Load(); s != 0 {
			t.Fatalf("state = %#x, want 0", s)
		}
	})
}

func TestOnceMutexLockContext(t *testing.T) {
	t.Run("DoneContextOnIdle", func(t *testing.T) {
		var m OnceMutex
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if leader, err := m.LockContext(ctx); leader || !errors.Is(err, context.Canceled) {
			t.Fatalf("LockContext = %v, %v; want false, context.Canceled", leader, err)
		}
		if !m.Lock() {
			t.Fatal("a canceled LockContext started the operation")
		}
		m.Unlock()
	})
	t.Run("DoneContextOnCompleted", func(t *testing.T) {
		var m OnceMutex
		m.Lock()
		m.Unlock()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if leader, err := m.LockContext(ctx); leader || err != nil {
			t.Fatalf("LockContext = %v, %v; want false, nil", leader, err)
		}
	})
	t.Run("Waiters", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var m OnceMutex
			if leader, err := m.LockContext(context.Background()); !leader || err != nil {
				t.Fatalf("first LockContext = %v, %v; want true, nil", leader, err)
			}
			canceled, cancel := context.WithCancel(context.Background())
			type result struct {
				leader bool
				err    error
			}
			waiting := make(chan result, 2)
			stopped := make(chan result, 1)
			for range 2 {
				go func() {
					leader, err := m.LockContext(context.Background())
					waiting <- result{leader, err}
				}()
			}
			go func() {
				leader, err := m.LockContext(canceled)
				stopped <- result{leader, err}
			}()
			synctest.Wait()
			cancel()
			if r := <-stopped; r.leader || !errors.Is(r.err, context.Canceled) {
				t.Fatalf("canceled waiter = %+v", r)
			}
			value := 42
			m.Unlock()
			for range 2 {
				if r := <-waiting; r.leader || r.err != nil {
					t.Fatalf("waiter = %+v, want false, nil", r)
				}
			}
			_ = value
		})
	})
}

func TestNamedOnceMutexLockContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var m NamedOnceMutex
		if leader, err := m.LockContext(context.Background(), "key"); !leader || err != nil {
			t.Fatalf("LockContext = %v, %v; want true, nil", leader, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if leader, err := m.LockContext(ctx, "key"); leader || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("waiter = %v, %v; want false, context.DeadlineExceeded", leader, err)
		}
		done := make(chan bool)
		go func() {
			leader, _ := m.LockContext(context.Background(), "key")
			done <- leader
		}()
		synctest.Wait()
		m.Unlock("key")
		if <-done {
			t.Fatal("a waiter became the leader of a completed operation")
		}
		ctx, cancel = context.WithCancel(context.Background())
		cancel()
		if leader, err := m.LockContext(ctx, "other"); leader || !errors.Is(err, context.Canceled) {
			t.Fatalf("done context = %v, %v; want false, context.Canceled", leader, err)
		}
		if !m.Lock("other") {
			t.Fatal("a canceled LockContext started an operation")
		}
		m.Unlock("other")
	})
}

func TestControlWaitGroupDoContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cwg := NewControlWaitGroup(1)
		release := make(chan struct{})
		if err := cwg.DoContext(context.Background(), func() { <-release }); err != nil {
			t.Fatalf("DoContext = %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		start := time.Now()
		if err := cwg.DoContext(ctx, func() { t.Error("a canceled task ran") }); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("DoContext = %v, want context.DeadlineExceeded", err)
		}
		if elapsed := time.Since(start); elapsed != time.Second {
			t.Errorf("waited %v, want 1s", elapsed)
		}
		if cwg.Waiting() != 0 {
			t.Fatal("a canceled submission is still counted as waiting")
		}
		if err := cwg.WaitContext(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("WaitContext with an expired context = %v", err)
		}
		waitCtx, cancelWait := context.WithTimeout(context.Background(), time.Second)
		defer cancelWait()
		if err := cwg.WaitContext(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("WaitContext on a busy group = %v, want context.DeadlineExceeded", err)
		}
		queued := make(chan error, 1)
		go func() { queued <- cwg.DoContext(context.Background(), func() {}) }()
		synctest.Wait()
		cwg.Abort()
		if err := <-queued; !errors.Is(err, ErrAborted) {
			t.Fatalf("DoContext after Abort = %v, want ErrAborted", err)
		}
		if err := cwg.DoContext(context.Background(), func() {}); !errors.Is(err, ErrAborted) {
			t.Fatalf("DoContext on an aborted group = %v, want ErrAborted", err)
		}
		close(release)
		if err := cwg.WaitContext(context.Background()); err != nil {
			t.Fatalf("WaitContext = %v", err)
		}
		if cwg.Working() != 0 || cwg.Waiting() != 0 {
			t.Fatal("tasks remain after WaitContext")
		}
	})
}
