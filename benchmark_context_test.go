//go:build !nsync_baseline

package nsync

import (
	"context"
	"testing"
)

// Context-aware acquisition did not exist before, so these benchmarks have no
// baseline counterpart. They measure the fast path and a contended loop that
// rarely parks; a parked wait with a cancelable context, which registers a
// context.AfterFunc callback, is not measured here.
func BenchmarkContext(b *testing.B) {
	background := context.Background()
	cancelable, cancel := context.WithCancel(background)
	defer cancel()
	for _, c := range []struct {
		name string
		ctx  context.Context
	}{{"Background", background}, {"Cancelable", cancelable}} {
		b.Run("TryMutex/"+c.name, func(b *testing.B) {
			m := NewTryMutex()
			b.ReportAllocs()
			for b.Loop() {
				if m.LockContext(c.ctx) != nil {
					b.Fatal("uncontended LockContext failed")
				}
				m.Unlock()
			}
		})
		b.Run("Semaphore/"+c.name, func(b *testing.B) {
			s := NewSemaphore(8)
			b.ReportAllocs()
			for b.Loop() {
				if s.AcquireContext(c.ctx) != nil {
					b.Fatal("uncontended AcquireContext failed")
				}
				s.Release()
			}
		})
		b.Run("NamedMutex/"+c.name, func(b *testing.B) {
			var m NamedMutex
			b.ReportAllocs()
			for b.Loop() {
				if m.LockContext(c.ctx, "key") != nil {
					b.Fatal("uncontended LockContext failed")
				}
				m.Unlock("key")
			}
		})
		b.Run("TryMutexContended/"+c.name, func(b *testing.B) {
			m := NewTryMutex()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if m.LockContext(c.ctx) != nil {
						b.Error("LockContext failed")
						return
					}
					m.Unlock()
				}
			})
		})
	}
}
