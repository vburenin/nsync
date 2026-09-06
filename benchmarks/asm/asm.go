//go:build amd64 || arm64

// Package asmbench compares handwritten assembly with compiler intrinsics.
// It is an isolated experiment, not a backend used by nsync.
package asmbench

//go:noescape
func acquire(ptr *int32) bool

//go:noescape
func release(ptr *int32) bool
