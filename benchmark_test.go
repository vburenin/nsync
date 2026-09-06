package nsync

import (
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Both the frozen baseline and the current implementation run these exact
// workloads. Concrete types keep interface dispatch out of the measurements.
// Run with -tags=nsync_baseline to select the frozen implementation.

func BenchmarkTryMutex(b *testing.B) {
	b.Run("Uncontended", func(b *testing.B) {
		m := newBenchTryMutex()
		b.ReportAllocs()
		for b.Loop() {
			m.Lock()
			m.Unlock()
		}
	})
	b.Run("TrySuccess", func(b *testing.B) {
		m := newBenchTryMutex()
		b.ReportAllocs()
		for b.Loop() {
			if !m.TryLock() {
				b.Fatal("uncontended TryLock failed")
			}
			m.Unlock()
		}
	})
	b.Run("TryFailure", func(b *testing.B) {
		m := newBenchTryMutex()
		m.Lock()
		defer m.Unlock()
		b.ReportAllocs()
		for b.Loop() {
			if m.TryLock() {
				b.Fatal("locked twice")
			}
		}
	})
	b.Run("TimeoutSuccess", func(b *testing.B) {
		m := newBenchTryMutex()
		b.ReportAllocs()
		for b.Loop() {
			if !m.TryLockTimeout(time.Second) {
				b.Fatal("uncontended timeout failed")
			}
			m.Unlock()
		}
	})
	b.Run("TimeoutExpired", func(b *testing.B) {
		m := newBenchTryMutex()
		m.Lock()
		defer m.Unlock()
		b.ReportAllocs()
		for b.Loop() {
			if m.TryLockTimeout(time.Nanosecond) {
				b.Fatal("locked twice")
			}
		}
	})
	b.Run("TimeoutPark", func(b *testing.B) {
		m := newBenchTryMutex()
		m.Lock()
		defer m.Unlock()
		b.ReportAllocs()
		for b.Loop() {
			if m.TryLockTimeout(10 * time.Microsecond) {
				b.Fatal("locked twice")
			}
		}
	})
	for _, work := range []int{0, 100} {
		b.Run(fmt.Sprintf("Contended%d", work), func(b *testing.B) {
			m := newBenchTryMutex()
			var value uint64
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					m.Lock()
					for range work {
						value = value*1664525 + 1013904223
					}
					m.Unlock()
				}
			})
			runtime.KeepAlive(value)
		})
	}
}

func BenchmarkSemaphore(b *testing.B) {
	for _, capacity := range []int{1, 8, 64} {
		b.Run(strconv.Itoa(capacity), func(b *testing.B) {
			b.Run("Uncontended", func(b *testing.B) {
				s := newBenchSemaphore(capacity)
				b.ReportAllocs()
				for b.Loop() {
					s.Acquire()
					s.Release()
				}
			})
			b.Run("Parallel", func(b *testing.B) {
				s := newBenchSemaphore(capacity)
				b.ReportAllocs()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						s.Acquire()
						s.Release()
					}
				})
			})
		})
	}
	b.Run("TryFailure", func(b *testing.B) {
		s := newBenchSemaphore(1)
		s.Acquire()
		defer s.Release()
		b.ReportAllocs()
		for b.Loop() {
			if s.TryAcquire() {
				b.Fatal("acquired full semaphore")
			}
		}
	})
	b.Run("TimeoutSuccess", func(b *testing.B) {
		s := newBenchSemaphore(1)
		b.ReportAllocs()
		for b.Loop() {
			if !s.TryAcquireTimeout(time.Second) {
				b.Fatal("acquisition failed")
			}
			s.Release()
		}
	})
	b.Run("Value", func(b *testing.B) {
		s := newBenchSemaphore(8)
		s.Acquire()
		var sum int
		b.ReportAllocs()
		for b.Loop() {
			sum += s.Value()
		}
		runtime.KeepAlive(sum)
	})
	b.Run("Full8TryFailure", func(b *testing.B) {
		s := newBenchSemaphore(8)
		for range 8 {
			s.Acquire()
		}
		b.ReportAllocs()
		for b.Loop() {
			if s.TryAcquire() {
				b.Fatal("acquired full semaphore")
			}
		}
	})
}

func BenchmarkNamedMutex(b *testing.B) {
	for _, length := range []int{32, 128, 1024} {
		b.Run(fmt.Sprintf("LongKey%d", length), func(b *testing.B) {
			m := newBenchNamedMutex()
			key := strings.Repeat("k", length)
			m.Lock(key)
			m.Unlock(key)
			b.ReportAllocs()
			for b.Loop() {
				m.Lock(key)
				m.Unlock(key)
			}
		})
	}
	b.Run("HotKey", func(b *testing.B) {
		m := newBenchNamedMutex()
		m.Lock("key")
		m.Unlock("key")
		b.ReportAllocs()
		for b.Loop() {
			m.Lock("key")
			m.Unlock("key")
		}
	})
	b.Run("SameKeyParallel", func(b *testing.B) {
		m := newBenchNamedMutex()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Lock("key")
				m.Unlock("key")
			}
		})
	})
	b.Run("IndependentKeys", func(b *testing.B) {
		m := newBenchNamedMutex()
		var id atomic.Int64
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			key := strconv.FormatInt(id.Add(1), 10)
			for pb.Next() {
				m.Lock(key)
				m.Unlock(key)
			}
		})
	})
	b.Run("ManyKeys", func(b *testing.B) {
		m := newBenchNamedMutex()
		var keys [1024]string
		for i := range keys {
			keys[i] = strconv.Itoa(i)
			m.Lock(keys[i])
			m.Unlock(keys[i])
		}
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			key := keys[i&1023]
			m.Lock(key)
			m.Unlock(key)
			i++
		}
	})
	b.Run("Create1024", func(b *testing.B) {
		var keys [1024]string
		for i := range keys {
			keys[i] = strconv.Itoa(i)
		}
		b.ReportAllocs()
		for b.Loop() {
			m := newBenchNamedMutex()
			for _, key := range keys {
				m.Lock(key)
				m.Unlock(key)
			}
		}
	})
}

func BenchmarkOnceMutex(b *testing.B) {
	b.Run("First", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var m benchOnceMutex
			if !m.Lock() {
				b.Fatal("first lock failed")
			}
			m.Unlock()
		}
	})
	b.Run("Completed", func(b *testing.B) {
		var m benchOnceMutex
		m.Lock()
		m.Unlock()
		b.ReportAllocs()
		for b.Loop() {
			if m.Lock() {
				b.Fatal("locked twice")
			}
		}
	})
	b.Run("CompletedParallel", func(b *testing.B) {
		var m benchOnceMutex
		m.Lock()
		m.Unlock()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if m.Lock() {
					b.Error("locked twice")
				}
			}
		})
	})
}

func BenchmarkNamedOnceMutex(b *testing.B) {
	b.Run("Uncontended", func(b *testing.B) {
		m := newBenchNamedOnceMutex()
		b.ReportAllocs()
		for b.Loop() {
			if m.Lock("key") {
				m.Unlock("key")
			}
		}
	})
	b.Run("SameKeyParallel", func(b *testing.B) {
		m := newBenchNamedOnceMutex()
		var leaders atomic.Int64
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var local int64
			for pb.Next() {
				if m.Lock("key") {
					local++
					m.Unlock("key")
				}
			}
			leaders.Add(local)
		})
		b.ReportMetric(float64(leaders.Load())/float64(b.N), "leaders/op")
	})
	b.Run("IndependentKeys", func(b *testing.B) {
		m := newBenchNamedOnceMutex()
		var id atomic.Int64
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var key any = id.Add(1)
			for pb.Next() {
				if m.Lock(key) {
					m.Unlock(key)
				}
			}
		})
	})
}

func BenchmarkTryFailureParallel(b *testing.B) {
	m := newBenchTryMutex()
	m.Lock()
	defer m.Unlock()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if m.TryLock() {
				b.Error("locked twice")
			}
		}
	})
}

func BenchmarkConstruction(b *testing.B) {
	b.Run("TryMutex", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			runtime.KeepAlive(newBenchTryMutex())
		}
	})
	b.Run("Semaphore", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			runtime.KeepAlive(newBenchSemaphore(8))
		}
	})
	b.Run("NamedMutexFirstKey", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			m := newBenchNamedMutex()
			m.Lock("key")
			m.Unlock("key")
		}
	})
	b.Run("NamedOnceFirstKey", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			m := newBenchNamedOnceMutex()
			m.Lock("key")
			m.Unlock("key")
		}
	})
}

func BenchmarkControlWaitGroup(b *testing.B) {
	for _, size := range []int{1, 16, 256} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			cwg := newBenchControlWaitGroup(size)
			b.ReportAllocs()
			for b.Loop() {
				cwg.Do(func() {})
			}
			cwg.Wait()
		})
	}
	b.Run("Aborted", func(b *testing.B) {
		cwg := newBenchControlWaitGroup(1)
		cwg.Abort()
		b.ReportAllocs()
		for b.Loop() {
			if cwg.Do(func() {}) {
				b.Fatal("admitted after Abort")
			}
		}
	})
}

func BenchmarkSyncFlag(b *testing.B) {
	b.Run("Read", func(b *testing.B) {
		var flag benchSyncFlag
		b.ReportAllocs()
		for b.Loop() {
			flag.IsSet()
		}
	})
	b.Run("Write", func(b *testing.B) {
		var flag benchSyncFlag
		b.ReportAllocs()
		for b.Loop() {
			flag.Set()
			flag.Unset()
		}
	})
}

func BenchmarkStandardMutex(b *testing.B) {
	b.Run("Uncontended", func(b *testing.B) {
		var m sync.Mutex
		b.ReportAllocs()
		for b.Loop() {
			m.Lock()
			m.Unlock()
		}
	})
	b.Run("Contended", func(b *testing.B) {
		var m sync.Mutex
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Lock()
				m.Unlock()
			}
		})
	})
}

// This diagnostic samples acquisition latency rather than inferring it from
// RunParallel throughput. The bounded rings retain recent samples; timing one
// acquisition in 64 reduces instrumentation overhead. Interface dispatch and
// sampling make its ns/op unsuitable for comparison to uninstrumented benches.
func BenchmarkWaitLatency(b *testing.B) {
	for _, work := range []int{0, 100} {
		b.Run(fmt.Sprintf("TryMutex/Work%d", work), func(b *testing.B) {
			benchmarkWaitLatency(b, newBenchTryMutex(), work)
		})
		b.Run(fmt.Sprintf("StandardMutex/Work%d", work), func(b *testing.B) {
			benchmarkWaitLatency(b, new(sync.Mutex), work)
		})
	}
}

func benchmarkWaitLatency(b *testing.B, m sync.Locker, work int) {
	var samplesMu sync.Mutex
	var samples []int64
	var value uint64
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		var ring [8192]int64
		var count, iteration int
		for pb.Next() {
			measure := iteration&63 == 0
			var start time.Time
			if measure {
				start = time.Now()
			}
			m.Lock()
			if measure {
				ring[count&8191] = int64(time.Since(start))
				count++
			}
			for range work {
				value = value*1664525 + 1013904223
			}
			m.Unlock()
			iteration++
		}
		samplesMu.Lock()
		samples = append(samples, ring[:min(count, len(ring))]...)
		samplesMu.Unlock()
	})
	b.StopTimer()
	slices.Sort(samples)
	if n := len(samples); n != 0 {
		b.ReportMetric(float64(samples[n/2]), "p50-wait-ns")
		b.ReportMetric(float64(samples[(n-1)*99/100]), "p99-wait-ns")
		b.ReportMetric(float64(samples[n-1]), "max-wait-ns")
		b.ReportMetric(float64(n), "samples")
	}
	runtime.KeepAlive(value)
}
