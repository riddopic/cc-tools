// Complete implementation of condition-based waiting utilities for Go tests.
// Context: Replaces arbitrary time.Sleep calls with condition-based polling.
//
// Three approaches shown:
// 1. waitForCondition - generic polling with timeout (for any boolean condition)
// 2. waitForEvent     - channel-based event waiting (for goroutine coordination)
// 3. waitForCount     - wait for accumulator to reach threshold
//
// Prefer require.Eventually from testify when it fits your use case.
package example_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// waitForCondition polls until condition returns true or timeout expires.
// Use when you need to wait for a state change that isn't channel-based.
func waitForCondition(t *testing.T, description string, timeout, interval time.Duration, condition func() bool) {
	t.Helper()

	deadline := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}

		select {
		case <-deadline:
			t.Fatalf("timeout waiting for %s after %v", description, timeout)
		case <-ticker.C:
			// try again
		}
	}
}

// waitForEvent waits for a value on a channel or times out.
// Use for goroutine coordination where results arrive via channels.
func waitForEvent[T any](t *testing.T, ch <-chan T, timeout time.Duration) T {
	t.Helper()

	select {
	case v := <-ch:
		return v
	case <-time.After(timeout):
		var zero T
		t.Fatalf("timeout waiting for event after %v", timeout)
		return zero
	}
}

// waitForCount waits until a counter reaches the target value.
// Use when waiting for N occurrences of an event.
func waitForCount(t *testing.T, counter *atomic.Int64, target int64, timeout time.Duration) {
	t.Helper()

	waitForCondition(t, "counter to reach target", timeout, 10*time.Millisecond, func() bool {
		return counter.Load() >= target
	})
}

// --- Usage examples ---

// BEFORE (flaky, 60% pass rate):
//
//   func TestProcessEvents_Flaky(t *testing.T) {
//       processor := NewProcessor()
//       go processor.Start()
//
//       time.Sleep(300 * time.Millisecond) // Hope processing starts in 300ms
//       processor.Submit(event1)
//       processor.Submit(event2)
//       time.Sleep(50 * time.Millisecond)  // Hope results arrive in 50ms
//
//       results := processor.Results()
//       require.Len(t, results, 2)         // Fails randomly
//   }

// AFTER (reliable, 100% pass rate):
//
//   func TestProcessEvents_Reliable(t *testing.T) {
//       processor := NewProcessor()
//       go processor.Start()
//
//       // Wait for processor to be ready (not arbitrary sleep)
//       require.Eventually(t, func() bool {
//           return processor.IsReady()
//       }, 5*time.Second, 10*time.Millisecond)
//
//       processor.Submit(event1)
//       processor.Submit(event2)
//
//       // Wait for results (not arbitrary sleep)
//       require.Eventually(t, func() bool {
//           return len(processor.Results()) >= 2
//       }, 5*time.Second, 10*time.Millisecond)
//
//       results := processor.Results()
//       require.Len(t, results, 2) // Always succeeds
//   }

// TestWaitForCondition demonstrates polling-based waiting.
func TestWaitForCondition(t *testing.T) {
	var ready atomic.Bool

	// Simulate async initialization
	go func() {
		time.Sleep(50 * time.Millisecond)
		ready.Store(true)
	}()

	waitForCondition(t, "ready flag", 5*time.Second, 10*time.Millisecond, func() bool {
		return ready.Load()
	})

	require.True(t, ready.Load())
}

// TestWaitForEvent demonstrates channel-based waiting.
func TestWaitForEvent(t *testing.T) {
	ch := make(chan string, 1)

	// Simulate async event
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- "done"
	}()

	result := waitForEvent(t, ch, 5*time.Second)
	require.Equal(t, "done", result)
}

// TestWaitForCount demonstrates counter-based waiting.
func TestWaitForCount(t *testing.T) {
	var counter atomic.Int64
	var wg sync.WaitGroup

	// Simulate 5 concurrent operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
			counter.Add(1)
		}()
	}

	waitForCount(t, &counter, 5, 5*time.Second)
	require.Equal(t, int64(5), counter.Load())

	wg.Wait()
}

// TestRequireEventually shows the preferred testify approach.
func TestRequireEventually(t *testing.T) {
	var ready atomic.Bool

	go func() {
		time.Sleep(50 * time.Millisecond)
		ready.Store(true)
	}()

	// require.Eventually is the simplest approach for most cases
	require.Eventually(t, func() bool {
		return ready.Load()
	}, 5*time.Second, 10*time.Millisecond, "expected ready flag to be set")
}
