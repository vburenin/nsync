//go:build !nsync_spin

package nsync

func (l *lockState) spin() bool { return false }
