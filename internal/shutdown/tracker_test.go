package shutdown

import (
	"testing"
	"time"
)

func TestTrackerWaitNoOps(t *testing.T) {
	t.Parallel()
	tr := NewTracker()
	if !tr.Wait(10 * time.Millisecond) {
		t.Fatal("Wait should return true immediately when no ops are tracked")
	}
}

func TestTrackerWaitForOps(t *testing.T) {
	t.Parallel()
	tr := NewTracker()

	tr.Add()
	go func() {
		time.Sleep(50 * time.Millisecond)
		tr.Done()
	}()

	if !tr.Wait(time.Second) {
		t.Fatal("Wait should return true after op completes")
	}
}

func TestTrackerWaitTimeout(t *testing.T) {
	t.Parallel()
	tr := NewTracker()

	tr.Add()
	// never call Done

	if tr.Wait(20 * time.Millisecond) {
		t.Fatal("Wait should return false on timeout")
	}

	tr.Done() // cleanup
}

func TestTrackerStopIdempotent(t *testing.T) {
	t.Parallel()
	tr := NewTracker()
	tr.Stop()
	tr.Stop() // should not panic
}

func TestTrackerMultipleOps(t *testing.T) {
	t.Parallel()
	tr := NewTracker()

	const n = 5
	for range n {
		tr.Add()
	}

	done := make(chan struct{})
	go func() {
		for range n {
			time.Sleep(10 * time.Millisecond)
			tr.Done()
		}
		close(done)
	}()

	if !tr.Wait(time.Second) {
		t.Fatal("Wait should return true after all ops complete")
	}

	<-done
}
