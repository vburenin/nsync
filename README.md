# nsync

[![Go Reference](https://pkg.go.dev/badge/github.com/vburenin/nsync.svg)](https://pkg.go.dev/github.com/vburenin/nsync)

Synchronization primitives for Go: timed locks, named locks, semaphores, a
bounded goroutine executor, and an atomic flag with an optional external lock.

Requires Go 1.27 or later. The development toolchain is Go 1.27.1.
The package uses only the standard library. Synchronization uses atomic fast
paths and condition variables, with no channels in the library implementation.
The default backend is portable Go; compiler intrinsics emit native atomic
instructions on ARM, x86, and other Go architectures.

```sh
go get github.com/vburenin/nsync
```

## TryMutex

Create a mutex with `NewTryMutex()`. It provides `Lock`, `Unlock`, `TryLock`, and
`TryLockTimeout(time.Duration)`. The try methods return whether the lock was
acquired. Unlocking an unlocked mutex panics.
Copies of a constructed `TryMutex` share the same lock.

Contended mutexes adaptively hand ownership to waiting goroutines after extended
waiting. They do not guarantee FIFO order or a strict acquisition-latency bound.

For ordinary locks without a timeout, the standard library's `sync.Mutex` also
provides `TryLock`.

## NamedMutex

`NamedMutex` acquires independent locks by string name. Its zero value is ready
to use; `NewNamedMutex()` is also available.

It provides `Lock(name)`, `Unlock(name)`, `TryLock(name)`, and
`TryLockTimeout(name, timeout)`. Unlocking an unknown or unlocked name panics.
Created locks are retained for the lifetime of the instance, so use a bounded
set of names.

## Semaphore

`NewSemaphore(capacity)` limits concurrent acquisitions. Capacity must be
positive; zero or negative capacities panic.

It provides `Acquire`, `Release`, `TryAcquire`, `TryAcquireTimeout`, and `Value`.
`Value` reports the number of occupied slots. Releasing without an acquisition
panics. Do not copy a semaphore after first use.

All timed acquisition methods try immediately before waiting. Zero or negative
timeouts perform a single nonblocking attempt.

## OnceMutex and NamedOnceMutex

`OnceMutex.Lock()` returns true for the first acquisition. Other calls block
until `Unlock`, then return false. Only the successful caller should unlock it.
Its zero value is ready to use.

`NamedOnceMutex` maintains an independent `OnceMutex` per key. Its zero value is
ready to use. Keys must be comparable and equal to themselves (avoid NaN keys).
`Unlock(key)` discards the completed mutex, so a later `Lock(key)` starts a new
cycle. Unlocking an unknown key does nothing. This can combine overlapping
cache refreshes for the same key into a single operation.

## ControlWaitGroup

`NewControlWaitGroup(poolSize)` limits the number of tasks running concurrently.
The pool size must be positive. `Do(func())` blocks until a slot is available,
then launches the task and returns true. `Wait()` waits for running tasks and
pending submissions to finish.

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

As with `sync.WaitGroup`, submission to an empty group must precede `Wait`.
Before reusing a group, wait for all previous `Wait` calls to return. `Working`
and `Waiting` are snapshots for monitoring, not synchronization.

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
```

See [performance measurements and design decisions](benchmarks/README.md) for
the frozen baseline, reproducible benchmarks, assembly experiments, allocation
measurements, and throughput/latency tradeoffs. Native results are included for
Apple M3 Max ARM64 and [AMD Ryzen 9 5950X on Linux](benchmarks/results/final-amd64/README.md).
