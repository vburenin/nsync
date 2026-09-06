//go:build !nsync_spin

package nsync

func (m *mutexState) spin() bool { return false }
