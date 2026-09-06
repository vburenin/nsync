//go:build !arm && !arm64 && !mips && !mipsle && !mips64 && !mips64le && !ppc64 && !ppc64le && !s390x

package nsync

const cacheLineSize = 64
