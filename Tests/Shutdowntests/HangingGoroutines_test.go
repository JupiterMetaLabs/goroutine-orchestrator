package Shutdowntests

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/App"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/Tests/Common"
)

func TestLocalManager_SafeShutdown_WithHangingGoroutines(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_SafeShutdown_WithHangingGoroutines ===")
	Common.ResetGlobalState()

	// Setup
	appMgr := App.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := Local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Spawn some normal goroutines that complete quickly
	fmt.Println("Spawning 3 normal goroutines...")
	var normalCompleted atomic.Int32
	for i := 0; i < 3; i++ {
		localMgr.Go("fast-worker", func(ctx context.Context) error {
			normalCompleted.Add(1)
			time.Sleep(100 * time.Millisecond)
			return nil
		}, Local.AddToWaitGroup("fast-worker"))
	}

	// Spawn hanging goroutines that run forever (but respect context cancellation)
	fmt.Println("Spawning 2 hanging goroutines...")
	var hangingCancelled atomic.Int32
	for i := 0; i < 2; i++ {
		localMgr.Go("hanging-worker", func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				hangingCancelled.Add(1)
				return ctx.Err()
			case <-time.After(1 * time.Hour):
				// This would run forever without cancellation
				return nil
			}
		}, Local.AddToWaitGroup("hanging-worker"))
	}

	fmt.Println("✓ All goroutines spawned (3 normal + 2 hanging)")

	// Safe shutdown - should complete fast workers, then cancel hanging ones
	fmt.Println("\nInitiating safe shutdown...")
	startTime := time.Now()
	err = localMgr.Shutdown(true)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
	fmt.Printf("✓ Safe shutdown completed in %v\n", elapsed)

	// Wait a bit for cancellations to propagate
	time.Sleep(100 * time.Millisecond)

	// Verify normal goroutines completed
	if normalCompleted.Load() != 3 {
		t.Errorf("Expected 3 normal goroutines to complete, got %d", normalCompleted.Load())
	} else {
		fmt.Println("✓ All 3 normal goroutines completed gracefully")
	}

	// Verify hanging goroutines were cancelled
	if hangingCancelled.Load() != 2 {
		t.Errorf("Expected 2 hanging goroutines to be cancelled, got %d", hangingCancelled.Load())
	} else {
		fmt.Println("✓ All 2 hanging goroutines were cancelled")
	}

	// Shutdown should not take too long (not 1 hour!)
	// Should be: 100ms (fast workers) + 5s (timeout) + some overhead
	maxExpectedTime := 7 * time.Second
	if elapsed > maxExpectedTime {
		t.Errorf("Shutdown took too long: %v (expected < %v)", elapsed, maxExpectedTime)
	} else {
		fmt.Printf("✓ Shutdown completed in reasonable time (%v)\n", elapsed)
	}
}

func TestLocalManager_SafeShutdown_AllHanging(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_SafeShutdown_AllHanging ===")
	Common.ResetGlobalState()

	// Setup
	appMgr := App.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := Local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Spawn only hanging goroutines
	fmt.Println("Spawning 5 hanging goroutines...")
	var cancelled atomic.Int32
	for i := 0; i < 5; i++ {
		localMgr.Go("infinite-loop", func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					cancelled.Add(1)
					return ctx.Err()
				case <-time.After(10 * time.Millisecond):
					// Keep looping
				}
			}
		}, Local.AddToWaitGroup("infinite-loop"))
	}
	fmt.Println("✓ All hanging goroutines spawned")

	// Safe shutdown - should timeout and force cancel
	fmt.Println("\nInitiating safe shutdown...")
	startTime := time.Now()
	err = localMgr.Shutdown(true)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
	fmt.Printf("✓ Safe shutdown completed in %v\n", elapsed)

	// Wait for cancellations
	time.Sleep(100 * time.Millisecond)

	// All should be cancelled
	if cancelled.Load() != 5 {
		t.Errorf("Expected 5 goroutines to be cancelled, got %d", cancelled.Load())
	} else {
		fmt.Println("✓ All 5 hanging goroutines were force-cancelled")
	}

	// Should complete quickly because ShutdownFunction cancels contexts
	// and these goroutines respect context cancellation
	maxExpectedTime := 1 * time.Second
	if elapsed > maxExpectedTime {
		t.Errorf("Shutdown took too long: %v (expected < %v)", elapsed, maxExpectedTime)
	} else {
		fmt.Printf("✓ Shutdown completed quickly (%v) - goroutines respected cancellation\n", elapsed)
	}
}
