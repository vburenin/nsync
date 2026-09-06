//go:build nsync_spin && (amd64 || arm64)

package nsync

func spinPause()
