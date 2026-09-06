package nsync

import (
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestCWaitGroup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cwg := NewControlWaitGroup(2)
		// Exercise reuse after Wait as well as queued and running tasks.
		for range 2 {
			release := make(chan struct{})
			task := func() { <-release }
			cwg.Do(task)
			cwg.Do(task)
			go cwg.Do(task)
			synctest.Wait()
			if got := cwg.Working(); got != 2 {
				t.Errorf("Working() = %d, want 2", got)
			}
			if got := cwg.Waiting(); got != 1 {
				t.Errorf("Waiting() = %d, want 1", got)
			}
			waited := make(chan struct{})
			go func() {
				cwg.Wait()
				close(waited)
			}()
			synctest.Wait()
			select {
			case <-waited:
				t.Error("Wait returned before tasks completed")
			default:
			}
			close(release)
			<-waited
			synctest.Wait()
			if cwg.Working() != 0 || cwg.Waiting() != 0 {
				t.Error("worker slots or pending tasks remain after Wait")
			}
		}
	})
}

func TestCWaitGroupConcurrentSubmitters(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const poolSize, tasks = 4, 64
		cwg := NewControlWaitGroup(poolSize)
		var active, completed atomic.Int32
		var submitters sync.WaitGroup
		for range tasks {
			submitters.Go(func() {
				if !cwg.Do(func() {
					if got := active.Add(1); got > poolSize {
						t.Errorf("%d tasks running, limit is %d", got, poolSize)
					}
					time.Sleep(time.Millisecond)
					active.Add(-1)
					completed.Add(1)
				}) {
					t.Error("task unexpectedly rejected")
				}
			})
		}
		submitters.Wait()
		cwg.Wait()
		if got := completed.Load(); got != tasks {
			t.Errorf("completed %d tasks, want %d", got, tasks)
		}
		if active.Load() != 0 || cwg.Working() != 0 || cwg.Waiting() != 0 {
			t.Error("tasks remain after Wait")
		}
	})
}

func TestAbortConcurrentWithSubmission(t *testing.T) {
	for range 50 {
		cwg := NewControlWaitGroup(2)
		start := make(chan struct{})
		var admitted, completed atomic.Int32
		var callers sync.WaitGroup
		for range 16 {
			callers.Go(func() {
				<-start
				if cwg.Do(func() { completed.Add(1) }) {
					admitted.Add(1)
				}
			})
		}
		callers.Go(func() {
			<-start
			cwg.Abort()
		})
		close(start)
		callers.Wait()
		cwg.Wait()
		if completed.Load() != admitted.Load() {
			t.Errorf("completed %d tasks, admitted %d", completed.Load(), admitted.Load())
		}
		if cwg.Working() != 0 || cwg.Waiting() != 0 {
			t.Fatal("worker slots or pending tasks remain after Abort and Wait")
		}
	}
}
