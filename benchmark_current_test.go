//go:build !nsync_baseline

package nsync

var (
	newBenchTryMutex         = NewTryMutex
	newBenchSemaphore        = NewSemaphore
	newBenchNamedMutex       = NewNamedMutex
	newBenchNamedOnceMutex   = NewNamedOnceMutex
	newBenchControlWaitGroup = NewControlWaitGroup
)

type benchOnceMutex = OnceMutex
type benchSyncFlag = SyncFlag
