//go:build amd64 || arm64

package asmbench

import (
	"sync/atomic"
	"testing"
)

func TestMachinePrimitives(t *testing.T) {
	var state int32
	if !acquire(&state) {
		t.Fatal("initial acquisition failed")
	}
	if acquire(&state) {
		t.Fatal("acquired twice")
	}
	if !release(&state) {
		t.Fatal("release failed")
	}
	if release(&state) {
		t.Fatal("idle release succeeded")
	}
	if !acquire(&state) {
		t.Fatal("reacquisition failed")
	}
	if atomic.LoadInt32(&state) != 1 {
		t.Fatal("assembly did not update state")
	}
	release(&state)
}

func BenchmarkAtomicPair(b *testing.B) {
	b.Run("Intrinsics", func(b *testing.B) {
		var state int32
		b.ReportAllocs()
		for b.Loop() {
			if !atomic.CompareAndSwapInt32(&state, 0, 1) {
				b.Fatal("acquisition failed")
			}
			if !atomic.CompareAndSwapInt32(&state, 1, 0) {
				b.Fatal("release failed")
			}
		}
	})
	b.Run("Assembly", func(b *testing.B) {
		var state int32
		b.ReportAllocs()
		for b.Loop() {
			if !acquire(&state) {
				b.Fatal("acquisition failed")
			}
			if !release(&state) {
				b.Fatal("release failed")
			}
		}
	})
}
