//go:build nsync_spin

package nsync

import "runtime"

func (l *lockState) spin() bool {
	if runtime.GOMAXPROCS(0) <= 1 {
		return false
	}
	for range 16 {
		if l.tryLock() {
			return true
		}
		spinPause()
	}
	return false
}
