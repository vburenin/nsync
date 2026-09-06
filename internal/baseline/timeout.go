// Frozen pre-optimization implementation. Used only by benchmark tests.
package baseline

import "time"

// acquireTimeout attempts an immediate acquisition before starting a timer.
// Otherwise an already expired timer could win even when capacity is available.
func acquireTimeout(c chan struct{}, timeout time.Duration) bool {
	select {
	case c <- struct{}{}:
		return true
	default:
	}
	if timeout <= 0 {
		return false
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case c <- struct{}{}:
		return true
	case <-timer.C:
		return false
	}
}
