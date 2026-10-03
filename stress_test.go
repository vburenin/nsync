package nsync

import (
	"context"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Each stress test runs with the default handoff threshold, and with handoff
// after any lost race, so most contended acquisitions are handed over.
func forEachHandoffMode(t *testing.T, f func(t *testing.T)) {
	for _, mode := range []struct {
		name      string
		threshold time.Duration
	}{{"Barging", 0}, {"Handoff", time.Nanosecond}} {
		t.Run(mode.name, func(t *testing.T) {
			handoffOverride.Store(int32(mode.threshold))
			defer handoffOverride.Store(0)
			f(t)
		})
	}
}

func stressRounds() int {
	if testing.Short() || raceEnabled {
		return 60
	}
	return 250
}

// acquireRandomly acquires through a randomly chosen method, including
// timeouts that expire while queued and contexts canceled concurrently.
func acquireRandomly(rng *rand.Rand, lock func(), try func() bool, timed func(time.Duration) bool, withContext func(context.Context) error) bool {
	switch rng.IntN(7) {
	case 0, 1:
		lock()
		return true
	case 2:
		return try()
	case 3:
		return timed(time.Duration(rng.IntN(3)) * time.Microsecond)
	case 4:
		return timed(time.Duration(rng.IntN(300)) * time.Microsecond)
	case 5:
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rng.IntN(200))*time.Microsecond)
		defer cancel()
		return withContext(ctx) == nil
	default:
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			runtime.Gosched()
			cancel()
		}()
		return withContext(ctx) == nil
	}
}

func TestStressTryMutex(t *testing.T) {
	forEachHandoffMode(t, func(t *testing.T) {
		m := NewTryMutex()
		var holders atomic.Int32
		var counter int
		var wg sync.WaitGroup
		for g := range 48 {
			wg.Go(func() {
				rng := rand.New(rand.NewPCG(uint64(g), 1))
				for range stressRounds() {
					if !acquireRandomly(rng, m.Lock, m.TryLock, m.TryLockTimeout, m.LockContext) {
						continue
					}
					if holders.Add(1) != 1 {
						t.Error("two goroutines hold the mutex")
					}
					counter++
					if rng.IntN(4) == 0 {
						runtime.Gosched()
					}
					holders.Add(-1)
					m.Unlock()
				}
			})
		}
		wg.Wait()
		if s := m.state.state.Load(); s != 0 {
			t.Fatalf("state = %#x after all goroutines left, want 0", s)
		}
	})
}

func TestStressSemaphore(t *testing.T) {
	forEachHandoffMode(t, func(t *testing.T) {
		for _, capacity := range []int{1, 2, 3, 8} {
			s := NewSemaphore(capacity)
			var holders atomic.Int32
			var wg sync.WaitGroup
			for g := range 48 {
				wg.Go(func() {
					rng := rand.New(rand.NewPCG(uint64(g), 2))
					for range stressRounds() {
						if !acquireRandomly(rng, s.Acquire, s.TryAcquire, s.TryAcquireTimeout, s.AcquireContext) {
							continue
						}
						if n := holders.Add(1); n > int32(capacity) {
							t.Errorf("%d holders, capacity %d", n, capacity)
						}
						if rng.IntN(4) == 0 {
							runtime.Gosched()
						}
						holders.Add(-1)
						s.Release()
					}
				})
			}
			wg.Wait()
			if st, gate := s.lock.state.Load(), s.gate.state.Load(); st != 0 || gate != 0 {
				t.Fatalf("capacity %d: state = %#x, gate = %#x after all goroutines left, want 0", capacity, st, gate)
			}
		}
	})
}

func TestStressNamedMutex(t *testing.T) {
	forEachHandoffMode(t, func(t *testing.T) {
		// Few names contend on cached locks; many names churn the shard
		// maps, their spares, and cache replacement.
		for _, names := range []int{1, 3, 200} {
			var m NamedMutex
			holders := make([]atomic.Int32, names)
			var wg sync.WaitGroup
			for g := range 48 {
				wg.Go(func() {
					rng := rand.New(rand.NewPCG(uint64(g), 3))
					for range stressRounds() {
						i := rng.IntN(names)
						name := strconv.Itoa(i)
						if !acquireRandomly(rng,
							func() { m.Lock(name) },
							func() bool { return m.TryLock(name) },
							func(d time.Duration) bool { return m.TryLockTimeout(name, d) },
							func(ctx context.Context) error { return m.LockContext(ctx, name) }) {
							continue
						}
						if holders[i].Add(1) != 1 {
							t.Errorf("two goroutines hold %q", name)
						}
						if rng.IntN(4) == 0 {
							runtime.Gosched()
						}
						holders[i].Add(-1)
						m.Unlock(name)
					}
				})
			}
			wg.Wait()
			checkNamedMutexIdle(t, &m)
		}
	})
}

// checkNamedMutexIdle verifies that no lock is in use and that idle locks are
// retained only within the documented bound.
func checkNamedMutexIdle(t *testing.T, m *NamedMutex) {
	t.Helper()
	if f := m.first.Load(); f != nil && f.lock.state.Load() != 0 {
		t.Errorf("first lock state = %#x, want 0", f.lock.state.Load())
	}
	table := m.table.Load()
	if table == nil {
		return
	}
	for i := range table {
		sh := table[i].Load()
		if sh == nil {
			continue
		}
		if n := sh.inUse(); n != 0 {
			t.Errorf("shard %d retains %d unlocked names", i, n)
		}
		if c := sh.cached.Load(); c != nil && c.lock.state.Load() != 0 {
			t.Errorf("shard %d cached lock state = %#x, want 0", i, c.lock.state.Load())
		}
		if sh.mu.state.Load() != 0 {
			t.Errorf("shard %d lock state = %#x, want 0", i, sh.mu.state.Load())
		}
	}
}

func TestStressNamedOnceMutex(t *testing.T) {
	forEachHandoffMode(t, func(t *testing.T) {
		var m NamedOnceMutex
		var leaders [5]atomic.Int32
		var started, completed [5]atomic.Int64
		var wg sync.WaitGroup
		for g := range 48 {
			wg.Go(func() {
				rng := rand.New(rand.NewPCG(uint64(g), 4))
				for range stressRounds() {
					key := rng.IntN(len(leaders))
					// A waiter joins the current or a later operation,
					// which completes before the waiter returns.
					before := started[key].Load()
					var leader bool
					var err error
					if rng.IntN(2) == 0 {
						leader = m.Lock(key)
					} else {
						ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rng.IntN(100))*time.Microsecond)
						leader, err = m.LockContext(ctx, key)
						cancel()
					}
					switch {
					case leader:
						started[key].Add(1)
						if leaders[key].Add(1) != 1 {
							t.Errorf("two leaders for key %d", key)
						}
						runtime.Gosched()
						completed[key].Add(1)
						leaders[key].Add(-1)
						m.Unlock(key)
					case err == nil && completed[key].Load() < before:
						t.Errorf("a waiter for key %d returned before the operation it joined completed", key)
					}
				}
			})
		}
		wg.Wait()
		shards := []*namedOnceShard{&m.primaryShard}
		if table := m.table.Load(); table != nil {
			for i := range table {
				if sh := table[i].Load(); sh != nil {
					shards = append(shards, sh)
				}
			}
		}
		for _, sh := range shards {
			if sh.primary != nil || sh.extraCount != 0 || len(sh.locks) != 0 {
				t.Error("completed operations were retained")
			}
		}
	})
}

func TestStressControlWaitGroup(t *testing.T) {
	forEachHandoffMode(t, func(t *testing.T) {
		for _, abort := range []bool{false, true} {
			cwg := NewControlWaitGroup(3)
			var running, admitted, completed atomic.Int32
			var wg sync.WaitGroup
			for g := range 32 {
				wg.Go(func() {
					rng := rand.New(rand.NewPCG(uint64(g), 5))
					for i := range stressRounds() / 4 {
						task := func() {
							if running.Add(1) > 3 {
								t.Error("more than 3 tasks running")
							}
							runtime.Gosched()
							running.Add(-1)
							completed.Add(1)
						}
						var ok bool
						if rng.IntN(2) == 0 {
							ok = cwg.Do(task)
						} else {
							ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rng.IntN(100))*time.Microsecond)
							ok = cwg.DoContext(ctx, task) == nil
							cancel()
						}
						if ok {
							admitted.Add(1)
						}
						if abort && g == 0 && i == 10 {
							cwg.Abort()
						}
					}
				})
			}
			wg.Wait()
			cwg.Wait()
			if completed.Load() != admitted.Load() {
				t.Errorf("completed %d tasks, admitted %d", completed.Load(), admitted.Load())
			}
			if s := cwg.state.Load(); s&^cwgAborted != 0 {
				t.Errorf("state = %#x after Wait, want no running tasks or flags", s)
			}
		}
	})
}
