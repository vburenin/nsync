// Package nsync provides timed and context-aware locks, named locks,
// semaphores, a bounded goroutine executor, and an atomic flag with an
// external lock.
//
// Acquisitions use atomic fast paths and park on condition variables, without
// channels. Waiters queue in arrival order; a released lock may still go to a
// running goroutine first, until a waiter has waited long enough to receive
// it directly. Timed and context-aware waits wake only the waiter that gives
// up.
//
// NamedMutex keeps a lock for each string name while it is held or awaited,
// plus at most 129 idle locks for reuse, so its memory does not grow with the
// number of names ever locked. NamedOnceMutex combines overlapping operations
// for a comparable key and removes completed operations. Semaphore bounds
// concurrent acquisitions, while ControlWaitGroup bounds concurrently running
// functions and supports canceling pending submissions.
//
// TryMutex, Semaphore, and ControlWaitGroup require their constructors. The
// zero values of NamedMutex, OnceMutex, NamedOnceMutex, and SyncFlag are usable.
package nsync
