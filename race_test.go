//go:build race

package nsync

// The race detector slows the stress tests severalfold; they run fewer rounds.
const raceEnabled = true
