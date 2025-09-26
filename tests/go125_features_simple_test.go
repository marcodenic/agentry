package tests

import (
	"context"
	"os"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func requireSynctest(t *testing.T) {
	t.Helper()
	if os.Getenv("AGENTRY_ENABLE_SYNCTEST") == "" {
		t.Skip("Set AGENTRY_ENABLE_SYNCTEST=1 to run synctest-based tests")
	}
}

// TestWaitGroupGoMethod demonstrates Go 1.25's new sync.WaitGroup.Go() method
func TestWaitGroupGoMethod(t *testing.T) {
	var results []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Go 1.25: Use the new WaitGroup.Go() method
	tasks := []string{"task1", "task2", "task3", "task4", "task5"}

	for _, task := range tasks {
		task := task // capture loop variable
		wg.Go(func() {
			// Simulate work
			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			results = append(results, task+" completed")
			mu.Unlock()
		})
	}

	wg.Wait()

	if len(results) != len(tasks) {
		t.Errorf("Expected %d results, got %d", len(tasks), len(results))
	}

	t.Logf("✅ WaitGroup.Go() method test passed with %d tasks", len(results))
}

// TestSynctestBasicFeatures demonstrates basic testing/synctest functionality
func TestSynctestBasicFeatures(t *testing.T) {
	requireSynctest(t)
	synctest.Test(t, func(t *testing.T) {
		var counter int
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Spawn multiple goroutines
		for i := 0; i < 10; i++ {
			wg.Go(func() {
				time.Sleep(100 * time.Millisecond) // This is virtualized in synctest
				mu.Lock()
				counter++
				mu.Unlock()
			})
		}

		wg.Wait()

		if counter != 10 {
			t.Errorf("Expected counter to be 10, got %d", counter)
		}

		t.Logf("✅ Synctest basic features test passed with counter: %d", counter)
	})
}

// TestSynctestWaitFunction demonstrates synctest.Wait() functionality
func TestSynctestWaitFunction(t *testing.T) {
	requireSynctest(t)
	synctest.Test(t, func(t *testing.T) {
		done := make(chan bool)

		go func() {
			time.Sleep(1 * time.Second) // Virtualized time
			done <- true
		}()

		// Wait for all goroutines to block
		synctest.Wait()

		// Receive the result
		select {
		case <-done:
			t.Log("✅ Synctest Wait() function test passed")
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for goroutine")
		}
	})
}

// TestChannelOperationsWithSynctest tests channel operations in synctest
func TestChannelOperationsWithSynctest(t *testing.T) {
	requireSynctest(t)
	synctest.Test(t, func(t *testing.T) {
		ch := make(chan int, 5)
		results := make([]int, 0, 10)
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Producer goroutine
		wg.Go(func() {
			for i := 0; i < 10; i++ {
				ch <- i
				time.Sleep(10 * time.Millisecond) // Virtualized
			}
			close(ch)
		})

		// Consumer goroutines
		for i := 0; i < 3; i++ {
			wg.Go(func() {
				for val := range ch {
					mu.Lock()
					results = append(results, val)
					mu.Unlock()
					time.Sleep(5 * time.Millisecond) // Virtualized
				}
			})
		}

		wg.Wait()

		if len(results) != 10 {
			t.Errorf("Expected 10 results, got %d", len(results))
		}

		t.Logf("✅ Channel operations test passed with %d results", len(results))
	})
}

// TestContextCancellationWithSynctest tests context cancellation in synctest
func TestContextCancellationWithSynctest(t *testing.T) {
	requireSynctest(t)
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		var wg sync.WaitGroup

		wg.Go(func() {
			select {
			case <-time.After(1 * time.Second): // This would timeout
				done <- false
			case <-ctx.Done():
				done <- true // Context should cancel first
			}
		})

		wg.Wait()

		select {
		case success := <-done:
			if !success {
				t.Error("Expected context cancellation, but got timeout")
			} else {
				t.Log("✅ Context cancellation test passed")
			}
		default:
			t.Error("No result received")
		}
	})
}
