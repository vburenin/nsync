package nsync

import (
	"sync"
	"testing"
)

func TestSyncFlag(t *testing.T) {
	var flag SyncFlag
	if flag.IsSet() || !flag.IsUnset() {
		t.Fatal("zero flag must be unset")
	}
	flag.Set()
	if !flag.IsSet() || flag.IsUnset() {
		t.Fatal("Set did not set the flag")
	}
	flag.Unset()
	if flag.IsSet() || !flag.IsUnset() {
		t.Fatal("Unset did not clear the flag")
	}
}

func TestSyncFlagConcurrent(t *testing.T) {
	var flag SyncFlag
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 100 {
				flag.Set()
				flag.IsSet()
				flag.Unset()
				flag.IsUnset()
			}
		})
	}
	for range 100 {
		flag.Lock()
		before := flag.IsSet()
		if after := flag.IsSet(); after != before {
			t.Error("flag changed while its mutex was held")
		}
		flag.Unlock()
	}
	wg.Wait()
	if flag.IsSet() {
		t.Error("flag must be unset after all workers finish with Unset")
	}
}
