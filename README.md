# nsync

[![Go Reference](https://pkg.go.dev/badge/github.com/vburenin/nsync.svg)](https://pkg.go.dev/github.com/vburenin/nsync)

Synchronization primitives for Go: timed and context-aware locks, named locks,
semaphores, a bounded goroutine executor, and an atomic flag with an optional
external lock.

Requires Go 1.27 or later. The development toolchain is Go 1.27.1.
The package uses only the standard library. Synchronization uses atomic fast
paths and condition variables, with no channels in the library implementation.
The default backend is portable Go. On amd64, arm64, the other 64-bit
architectures, and 32-bit MIPS, compiler intrinsics emit the atomic
instructions inline; on 386, 32-bit ARM, and WebAssembly, `sync/atomic`
operations are calls into the runtime's atomic routines.

```sh
go get github.com/vburenin/nsync
```

## TryMutex

Create a mutex with `NewTryMutex()`. It provides `Lock`, `Unlock`, `TryLock`,
`TryLockTimeout(time.Duration)`, and `LockContext(ctx)`. The try methods return
whether the lock was acquired. Unlocking an unlocked mutex panics.
Copies of a constructed `TryMutex` share the same lock.

Waiters queue in arrival order. A released mutex may be taken by a running
goroutine before the woken waiter runs, which keeps contended throughput high;
once the first waiter has waited about 100 µs, ownership is handed to waiters
directly. This is not a strict FIFO or latency guarantee.

For ordinary locks without a timeout, the standard library's `sync.Mutex` also
provides `TryLock`.

## NamedMutex

`NamedMutex` acquires independent locks by string name. Its zero value is ready
to use; `NewNamedMutex()` is also available.

It provides `Lock(name)`, `Unlock(name)`, `TryLock(name)`,
`TryLockTimeout(name, timeout)`, and `LockContext(ctx, name)`. Unlocking a name
that is not locked panics.

Locks for names that no goroutine holds or waits for are discarded, apart from
the bounded set of idle locks described next, so any number of distinct names
can be used: memory follows the names currently in use, not every name ever
locked. To keep repeated locking of hot names allocation-free,
an instance retains at most 129 idle locks: one for the first name it locked,
and up to one cached and one spare lock in each of its 64 shards.

## Semaphore

`NewSemaphore(capacity)` limits concurrent acquisitions. Capacity must be
positive; zero or negative capacities panic. On 32-bit platforms, capacities
above 67,108,863 are reduced to it.

It provides `Acquire`, `Release`, `TryAcquire`, `TryAcquireTimeout`,
`AcquireContext(ctx)`, and `Value`. `Value` reports the number of occupied
slots. Releasing without an acquisition panics. Do not copy a semaphore after
first use.

All timed acquisition methods try immediately before waiting. Zero or negative
timeouts perform a single nonblocking attempt.

## OnceMutex and NamedOnceMutex

`OnceMutex.Lock()` returns true for the first acquisition. Other calls block
until `Unlock`, then return false. Only the successful caller should unlock it.
Its zero value is ready to use. `LockContext(ctx)` also stops waiting when the
context is done.

`NamedOnceMutex` maintains an independent `OnceMutex` per key. Its zero value is
ready to use. Keys must be comparable and equal to themselves (avoid NaN keys).
`Unlock(key)` discards the completed mutex, so a later `Lock(key)` starts a new
cycle. Unlocking an unknown key does nothing. This can combine overlapping
cache refreshes for the same key into a single operation.
`LockContext(ctx, key)` is the context-aware form of `Lock`.

## ControlWaitGroup

`NewControlWaitGroup(poolSize)` limits the number of tasks running concurrently.
The pool size must be positive. `Do(func())` blocks until a slot is available,
then launches the task and returns true. `Wait()` waits for running tasks and
pending submissions to finish. Blocked submissions are admitted in arrival order.

```go
workers := nsync.NewControlWaitGroup(4)
for _, job := range jobs {
    workers.Do(func() { process(job) })
}
workers.Wait()
```

`Abort()` permanently rejects new tasks and unblocks pending `Do` calls, which
return false. Already admitted tasks may continue running; call `Wait` to wait
for them. Repeated calls to `Abort` are safe.

`DoContext(ctx, func())` and `WaitContext(ctx)` are the context-aware forms.
`DoContext` returns `ErrAborted` once the group is aborted, unless its
context is already done when it is called.

As with `sync.WaitGroup`, submission to an empty group must precede `Wait`.
Before reusing a group, wait for all previous `Wait` calls to return. `Working`
and `Waiting` are snapshots for monitoring, not synchronization.

## Context-aware waiting

`TryMutex.LockContext`, `NamedMutex.LockContext`, `Semaphore.AcquireContext`,
and `ControlWaitGroup.DoContext` return nil once they acquire. Otherwise they
return `ctx.Err()` without acquiring, except that `DoContext` called with a
live context returns `ErrAborted` once the group is aborted. As in
`golang.org/x/sync/semaphore`, a context that is already done fails
immediately, even if the lock is free. `OnceMutex.LockContext` and
`ControlWaitGroup.WaitContext` still report an operation or group that has
already completed.

The context is checked once before the first attempt and otherwise only while
waiting. On an AMD Ryzen 9 5950X, `context.Background()` adds about 0.15 ns to
an uncontended `TryMutex.LockContext` (a `Lock`/`Unlock` pair takes 3.6 ns
without a context), under 0.1 ns to `Semaphore.AcquireContext`, and about 1.5
ns to `NamedMutex.LockContext` (6.2 ns per pair); a cancelable context adds
0.4–0.6 ns, or about 2.2 ns for `NamedMutex`. A canceled or timed-out waiter
wakes alone and leaves the queue without disturbing other waiters.

## SyncFlag

`SyncFlag` has an unset zero value. `Set` and `Unset` serialize writes with its
embedded mutex; `IsSet` and `IsUnset` read atomically. Hold `Lock` to prevent
changes, then call `Unlock` to allow changes again. Calling `Set` or `Unset`
while holding that mutex would deadlock.

## Development

```sh
go test -race -shuffle=on ./...
go vet ./...
./benchmarks/cross-build.sh
./benchmarks/compare-revisions.sh <git-revision> benchmarks/results/local
```

Stress tests mix blocking, nonblocking, timed, and context-aware acquisitions,
with and without forced handoff, and check exclusion, capacity, and that every
lock returns to its idle state. Use `-short` for a quicker run.

See [performance measurements and design decisions](benchmarks/README.md) for
the design, reproducible benchmarks, rejected experiments, allocation
measurements, and throughput/latency tradeoffs. The current code is measured in
the [round 3 report](benchmarks/results/round3-amd64/README.md) on an AMD Ryzen 9
5950X; earlier rounds include Apple M3 Max ARM64 results.
