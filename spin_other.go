//go:build nsync_spin && !amd64 && !arm64

package nsync

// Other architectures use the bounded loop's atomic loads as polling points.
func spinPause() {}
