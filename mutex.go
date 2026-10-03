package nsync

import (
	"context"
	"sync/atomic"
	"time"
)

// lockState is the core of TryMutex, NamedMutex, and Semaphore. One atomic
// word holds the number of held permits and the wait-queue flags; a mutex is a
// lockState with a limit of one.
//
// An uncontended acquisition or release is a single compare-and-swap.
// Contended acquirers park in FIFO order on a lazily allocated queue, each on
// its own condition variable. A release that finds stateQueued set wakes the
// first waiter and clears the flag; the woken waiter becomes responsible for
// setting it again if waiters still need a wakeup. Until then, running
// goroutines keep using the fast paths, and they may take a free permit before
// the woken waiter runs. A waiter that loses that race keeps its place at the
// head of the queue. Once it has waited longer than handoffThreshold, releases
// hand permits to the head directly, so newcomers cannot keep overtaking it.
type lockState struct {
	state atomic.Uintptr
	queue atomic.Pointer[waitQueue]
}

// lockWord is the state word, as wide as a pointer. On 32-bit platforms Go
// implements 64-bit atomic operations with function calls, and on 32-bit MIPS
// with a lock shared by the whole process.
type lockWord = uintptr

const lockWordBits = 32 << (^lockWord(0) >> 63) // 32 or 64

// The permit count sits between the flags. Starving and retired states are
// above any count, so state < limit<<stateUnitShift alone means that a permit
// is free and may be taken, with or without queued waiters.
const (
	stateQueued    lockWord = 1      // waiters need a wakeup: releases take the slow path
	stateUnitShift          = 3      // low bits are reserved for flags
	stateUnit      lockWord = 1 << 3 // one held permit
	stateStarving  lockWord = 1 << (lockWordBits - 2)
	stateRetired   lockWord = 1 << (lockWordBits - 1) // a cached NamedMutex entry was discarded; look it up again

	// stateNegative marks a release without a held permit. It sits directly
	// above the count, so any decrement below zero sets it, including one
	// that borrows from stateStarving instead of wrapping around.
	stateNegative lockWord = 1 << (lockWordBits - 3)

	// maxPermits bounds semaphore capacities so the count stays below the
	// flags: about 2.9e17, or 6.7e7 on 32-bit platforms.
	maxPermits = stateNegative>>stateUnitShift - 1
)

// handoffThreshold is how long the queue head may keep losing races to running
// goroutines before permits are handed to it directly. It is a scheduling
// heuristic, not a bound on how long an acquisition can take.
const handoffThreshold = 100 * time.Microsecond

// handoffOverride, in nanoseconds, replaces handoffThreshold when non-zero;
// tests lower it to exercise handoff. It is 32-bit, which limits it to about
// 2.1 s, because 32-bit MIPS implements 64-bit atomic operations with a lock
// shared by the process.
var handoffOverride atomic.Int32

func starved(since time.Time) bool {
	threshold := handoffThreshold
	if d := handoffOverride.Load(); d != 0 {
		threshold = time.Duration(d)
	}
	return time.Since(since) >= threshold
}

type acquireResult uint8

const (
	acquired acquireResult = iota
	canceled               // the timeout expired or the context was done
	retired                // the NamedMutex entry was discarded
)

func (l *lockState) lock() {
	if !l.state.CompareAndSwap(0, stateUnit) {
		l.lockSlow()
	}
}

//go:noinline
func (l *lockState) lockSlow() {
	if !l.spin() {
		l.acquireSlow(1, 0, nil)
	}
}

// tryLock may take a free lock ahead of queued waiters, as Lock does, but not
// while the queue is starving. A failed attempt does not write.
func (l *lockState) tryLock() bool {
	s := l.state.Load()
	return s <= stateQueued && l.state.CompareAndSwap(s, s+stateUnit)
}

func (l *lockState) lockTimeout(timeout time.Duration) bool {
	return l.state.CompareAndSwap(0, stateUnit) || l.lockTimeoutSlow(timeout)
}

//go:noinline
func (l *lockState) lockTimeoutSlow(timeout time.Duration) bool {
	return timeout > 0 && l.acquireSlow(1, timeout, nil) == acquired
}

func (l *lockState) lockContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if l.state.CompareAndSwap(0, stateUnit) || l.acquireSlow(1, 0, ctx) == acquired {
		return nil
	}
	return ctx.Err()
}

func (l *lockState) unlock() {
	if !l.state.CompareAndSwap(stateUnit, 0) {
		l.unlockSlow()
	}
}

//go:noinline
func (l *lockState) unlockSlow() {
	if !l.releaseSlow(1) {
		panic("nsync: unlock of unlocked mutex")
	}
}

// tryAcquire takes a free permit unless the queue is starving. Unlike
// tryLock, it retries when a concurrent update changes the count.
func (l *lockState) tryAcquire(limit lockWord) bool {
	for {
		s := l.state.Load()
		if s >= limit<<stateUnitShift {
			return false
		}
		if l.state.CompareAndSwap(s, s+stateUnit) {
			return true
		}
	}
}

// permits returns the number of held permits in a state word.
func permits(s lockWord) lockWord { return s & (stateStarving - 1) >> stateUnitShift }

// retiredWord reports whether a state word belongs to a discarded NamedMutex
// entry. A negative word, whose count fell below zero, is instead a semaphore
// release without an acquisition that is about to be undone.
func retiredWord(s lockWord) bool { return s&(stateRetired|stateNegative) == stateRetired }

// clearable returns the flags that may be cleared from s. In a negative word
// the starving bit belongs to the wrapped count until the release is undone.
func clearable(s lockWord) lockWord {
	if s&stateNegative != 0 {
		return stateQueued
	}
	return stateQueued | stateStarving
}

func (l *lockState) waitQueue() *waitQueue {
	if q := l.queue.Load(); q != nil {
		return q
	}
	q := new(waitQueue)
	if l.queue.CompareAndSwap(nil, q) {
		return q
	}
	return l.queue.Load()
}

// acquireSlow acquires one of limit permits after the fast path failed. A
// positive timeout or a non-nil ctx bounds the wait.
func (l *lockState) acquireSlow(limit lockWord, timeout time.Duration, ctx context.Context) acquireResult {
	var start time.Time
	if timeout > 0 {
		start = time.Now()
	}
	// Barge while a permit is free: queued waiters only hold back newcomers
	// once the queue is starving.
	for {
		s := l.state.Load()
		if s < limit<<stateUnitShift {
			if l.state.CompareAndSwap(s, s+stateUnit) {
				return acquired
			}
			continue
		}
		if retiredWord(s) {
			return retired
		}
		break
	}
	if timeout > 0 {
		// A tiny timeout can expire before parking; no timer is needed then.
		if timeout -= time.Since(start); timeout <= 0 {
			return canceled
		}
	} else if ctx != nil && ctx.Err() != nil {
		return canceled
	}
	q := l.waitQueue()
	q.mu.Lock()
	// Releases cannot miss this waiter: once stateQueued is set they take
	// q.mu, and until then their compare-and-swap fails on the changed word.
	for {
		s := l.state.Load()
		if retiredWord(s) {
			q.mu.Unlock()
			return retired
		}
		if s < limit<<stateUnitShift {
			if l.state.CompareAndSwap(s, s+stateUnit) {
				q.mu.Unlock()
				return acquired
			}
			continue
		}
		if s&stateQueued != 0 || l.state.CompareAndSwap(s, s|stateQueued) {
			break
		}
	}
	w := q.take()
	if w.since = start; timeout <= 0 {
		w.since = time.Now()
	}
	q.pushBack(w)
	c := w.arm(timeout, ctx)
	result := l.park(w, q, limit)
	w.disarm(c)
	q.keep(w)
	q.mu.Unlock()
	return result
}

// park waits with q.mu held until w acquires a permit or is canceled.
func (l *lockState) park(w *waiter, q *waitQueue, limit lockWord) acquireResult {
	for {
		w.cond.Wait()
		switch {
		case w.granted:
			// A waiter that did not wait long shows the queue is moving:
			// let running goroutines take free permits again.
			if !starved(w.since) {
				l.clear(stateStarving)
			}
			return acquired
		case w.canceled:
			l.leave(w, q, limit)
			return canceled
		case w.woken:
			if l.retry(w, q, limit) {
				return acquired
			}
		}
	}
}

// retry lets a woken waiter take a free permit. If a running goroutine got
// there first, the waiter parks again at its position and, once it has waited
// too long, makes releases hand permits to the queue head. Either way, it
// re-arms stateQueued while waiters remain.
func (l *lockState) retry(w *waiter, q *waitQueue, limit lockWord) bool {
	w.woken = false
	starving := starved(w.since)
	for {
		s := l.state.Load()
		if s < limit<<stateUnitShift {
			ns := s + stateUnit
			if q.head == w && w.next == nil {
				ns &^= stateQueued
			} else {
				ns |= stateQueued
			}
			if l.state.CompareAndSwap(s, ns) {
				q.remove(w)
				// Permits freed by several releases may have woken only
				// this waiter; pass the wakeup on.
				if limit > 1 && ns < limit<<stateUnitShift {
					q.wakeOne()
				}
				return true
			}
			continue
		}
		ns := s | stateQueued
		// In a negative word the starving bit belongs to the wrapped count.
		if starving && s&stateNegative == 0 {
			ns |= stateStarving
		}
		if ns == s || l.state.CompareAndSwap(s, ns) {
			return false
		}
	}
}

// leave runs after the waiter's callback removed it from the queue. A woken
// waiter passes its responsibility for the queue on.
func (l *lockState) leave(w *waiter, q *waitQueue, limit lockWord) {
	for {
		s := l.state.Load()
		ns := s
		wake := false
		if q.head == nil {
			ns &^= clearable(s)
		} else if w.woken {
			if s < limit<<stateUnitShift {
				wake = true
			} else {
				ns |= stateQueued
			}
		}
		if ns == s || l.state.CompareAndSwap(s, ns) {
			if wake {
				q.wakeOne()
			}
			return
		}
	}
}

func (l *lockState) clear(flag lockWord) {
	for {
		s := l.state.Load()
		if s&flag&clearable(s) == 0 || l.state.CompareAndSwap(s, s&^(flag&clearable(s))) {
			return
		}
	}
}

// releaseSlow releases a permit after the fast path failed. It reports false,
// changing nothing, if no permit is held.
func (l *lockState) releaseSlow(limit lockWord) bool {
	for {
		s := l.state.Load()
		if permits(s) == 0 || s&stateRetired != 0 {
			return false
		}
		if s&(stateQueued|stateStarving) != 0 {
			break
		}
		if l.state.CompareAndSwap(s, s-stateUnit) {
			return true
		}
	}
	q := l.queue.Load()
	q.mu.Lock()
	for {
		s := l.state.Load()
		h := q.head
		if s&stateStarving != 0 && h != nil {
			// Hand the permit over without releasing it.
			ns := s
			if h.next == nil {
				ns &^= stateQueued | stateStarving
			}
			if ns == s || l.state.CompareAndSwap(s, ns) {
				q.remove(h)
				h.granted = true
				h.cond.Signal()
				break
			}
			continue
		}
		// Wake one waiter, which takes over stateQueued. A mutex waiter that
		// is already awake will retry anyway.
		var next *waiter
		if h != nil && (limit > 1 || !h.woken) {
			for next = h; next != nil && next.woken; next = next.next {
			}
		}
		if l.state.CompareAndSwap(s, (s-stateUnit)&^(stateQueued|stateStarving)) {
			if next != nil {
				next.woken = true
				next.cond.Signal()
			}
			break
		}
	}
	q.mu.Unlock()
	return true
}

// releaseAdded finishes a release whose decrement was already applied and
// found waiters to wake or a starving queue.
func (l *lockState) releaseAdded(limit lockWord) {
	q := l.queue.Load()
	q.mu.Lock()
	for {
		s := l.state.Load()
		h := q.head
		if s&stateStarving != 0 && h != nil {
			// While the queue starves, a free permit belongs to its head.
			// The decrement may have happened before starvation began, so
			// a newcomer may have taken the permit; its release hands over.
			if permits(s) >= limit {
				break
			}
			ns := s + stateUnit
			if h.next == nil {
				ns &^= stateQueued | stateStarving
			}
			if l.state.CompareAndSwap(s, ns) {
				q.remove(h)
				h.granted = true
				h.cond.Signal()
				break
			}
			continue
		}
		var next *waiter
		if h != nil && (limit > 1 || !h.woken) {
			for next = h; next != nil && next.woken; next = next.next {
			}
		}
		if l.state.CompareAndSwap(s, s&^clearable(s)) {
			if next != nil {
				next.woken = true
				next.cond.Signal()
			}
			break
		}
	}
	q.mu.Unlock()
}
