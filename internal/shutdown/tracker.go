package shutdown

import (
	"sync"
	"time"
)

// Tracker coordinates graceful shutdown by tracking in-flight critical
// operations so the shutdown sequence can wait for them to finish before
// closing resources.
type Tracker struct {
	stopping chan struct{}
	wg       sync.WaitGroup
	once     sync.Once
}

// NewTracker creates a ready-to-use shutdown tracker.
func NewTracker() *Tracker {
	return &Tracker{stopping: make(chan struct{})}
}

// Add marks that a critical operation has started. Call Done when it finishes.
func (t *Tracker) Add() {
	t.wg.Add(1)
}

// Done marks a critical operation as finished.
func (t *Tracker) Done() {
	t.wg.Done()
}

// Stop signals that shutdown has begun. Safe to call multiple times.
func (t *Tracker) Stop() {
	t.once.Do(func() { close(t.stopping) })
}

// Wait blocks until all tracked operations complete or the timeout elapses.
// Returns true if all operations finished, false on timeout.
func (t *Tracker) Wait(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
