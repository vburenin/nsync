package nsync

import (
	"testing"
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
