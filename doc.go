// Package nsync provides timed locks, named locks, semaphores, a bounded
// goroutine executor, and an atomic flag with an external lock.
//
// Acquisitions use atomic fast paths and condition variables for parking,
// without channels. Contended mutexes adaptively hand ownership to waiting
// goroutines to limit repeated barging; FIFO order is not guaranteed.
//
// NamedMutex retains a lock for each string name. NamedOnceMutex combines
// overlapping operations for a comparable key and removes completed operations.
// Semaphore bounds concurrent acquisitions, while ControlWaitGroup bounds
// concurrently running functions and supports canceling pending submissions.
//
// TryMutex, Semaphore, and ControlWaitGroup require their constructors. The
// zero values of NamedMutex, OnceMutex, NamedOnceMutex, and SyncFlag are usable.
package nsync
