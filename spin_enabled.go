//go:build nsync_spin

package nsync

import "runtime"

func (m *mutexState) spin() bool {
	if runtime.GOMAXPROCS(0) <= 1 {
		return false
	}
	for range 16 {
		if m.tryLock() {
			return true
		}
		spinPause()
	}
	return false
}
