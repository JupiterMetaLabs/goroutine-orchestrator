package Managertests

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

// TestGo_Basic tests Go() without any options
func TestGo_Basic(t *testing.T) {
	fmt.Println("\n=== TestGo_Basic ===")
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

	// Spawn goroutine without options
	var completed atomic.Bool
	err = localMgr.Go("basic-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		completed.Store(true)
		return nil
	})

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait a bit for goroutine to complete
	time.Sleep(100 * time.Millisecond)

	if !completed.Load() {
		t.Error("Goroutine should have completed")
	}

	// Verify routine count
	count := localMgr.GetGoroutineCount()
	if count != 0 {
		t.Errorf("Expected 0 routines after completion, got %d", count)
	}

	fmt.Println("✓ Basic Go() test passed")
}

// TestGo_WithTimeout tests Go() with WithTimeout option
func TestGo_WithTimeout(t *testing.T) {
	fmt.Println("\n=== TestGo_WithTimeout ===")
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

	// Test 1: Goroutine completes before timeout
	var completed1 atomic.Bool
	err = localMgr.Go("fast-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		completed1.Store(true)
		return nil
	}, Local.WithTimeout(200*time.Millisecond))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if !completed1.Load() {
		t.Error("Fast goroutine should have completed")
	}

	// Test 2: Goroutine times out
	var timedOut atomic.Bool
	err = localMgr.Go("slow-worker", func(ctx context.Context) error {
		// Wait longer than timeout
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			timedOut.Store(true)
			return ctx.Err()
		}
	}, Local.WithTimeout(100*time.Millisecond))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for timeout to occur
	time.Sleep(150 * time.Millisecond)
	if !timedOut.Load() {
		t.Error("Slow goroutine should have timed out")
	}

	fmt.Println("✓ WithTimeout test passed")
}

// TestGo_WithPanicRecovery tests Go() with WithPanicRecovery option
func TestGo_WithPanicRecovery(t *testing.T) {
	fmt.Println("\n=== TestGo_WithPanicRecovery ===")
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

	// Test 1: Panic recovery enabled - panic should be caught
	err = localMgr.Go("panic-worker", func(ctx context.Context) error {
		panic("test panic")
	}, Local.WithPanicRecovery(true))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for panic to be recovered
	time.Sleep(100 * time.Millisecond)

	// Verify routine was cleaned up (panic recovery should allow cleanup)
	count := localMgr.GetGoroutineCount()
	if count != 0 {
		t.Errorf("Expected 0 routines after panic recovery, got %d", count)
	}

	// Test 2: Panic recovery disabled - should still complete (panic in goroutine doesn't crash test)
	err = localMgr.Go("no-recovery-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}, Local.WithPanicRecovery(false))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	fmt.Println("✓ WithPanicRecovery test passed")
}

// TestGo_AddToWaitGroup tests Go() with AddToWaitGroup option
func TestGo_AddToWaitGroup(t *testing.T) {
	fmt.Println("\n=== TestGo_AddToWaitGroup ===")
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

	// Spawn multiple goroutines with wait group
	var counter atomic.Int32
	numGoroutines := 5

	for i := 0; i < numGoroutines; i++ {
		err = localMgr.Go("waitgroup-worker", func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			counter.Add(1)
			return nil
		}, Local.AddToWaitGroup("waitgroup-worker"))

		if err != nil {
			t.Fatalf("Go() failed: %v", err)
		}
	}

	// Wait for all goroutines to complete using wait group
	err = localMgr.WaitForFunction("waitgroup-worker")
	if err != nil {
		t.Fatalf("WaitForFunction() failed: %v", err)
	}

	// Verify all completed
	if counter.Load() != int32(numGoroutines) {
		t.Errorf("Expected %d goroutines to complete, got %d", numGoroutines, counter.Load())
	}

	fmt.Println("✓ AddToWaitGroup test passed")
}

// TestGo_WithTimeoutAndPanicRecovery tests Go() with both timeout and panic recovery
func TestGo_WithTimeoutAndPanicRecovery(t *testing.T) {
	fmt.Println("\n=== TestGo_WithTimeoutAndPanicRecovery ===")
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

	// Test: Timeout + Panic Recovery
	var completed atomic.Bool
	err = localMgr.Go("timeout-panic-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		completed.Store(true)
		return nil
	}, Local.WithTimeout(200*time.Millisecond), Local.WithPanicRecovery(true))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if !completed.Load() {
		t.Error("Goroutine should have completed")
	}

	fmt.Println("✓ WithTimeout + WithPanicRecovery test passed")
}

// TestGo_WithTimeoutAndWaitGroup tests Go() with timeout and wait group
func TestGo_WithTimeoutAndWaitGroup(t *testing.T) {
	fmt.Println("\n=== TestGo_WithTimeoutAndWaitGroup ===")
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

	// Spawn goroutines with timeout and wait group
	var counter atomic.Int32
	numGoroutines := 3

	for i := 0; i < numGoroutines; i++ {
		err = localMgr.Go("timeout-wg-worker", func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			counter.Add(1)
			return nil
		}, Local.WithTimeout(200*time.Millisecond), Local.AddToWaitGroup("timeout-wg-worker"))

		if err != nil {
			t.Fatalf("Go() failed: %v", err)
		}
	}

	// Wait for all to complete
	err = localMgr.WaitForFunction("timeout-wg-worker")
	if err != nil {
		t.Fatalf("WaitForFunction() failed: %v", err)
	}

	if counter.Load() != int32(numGoroutines) {
		t.Errorf("Expected %d goroutines to complete, got %d", numGoroutines, counter.Load())
	}

	fmt.Println("✓ WithTimeout + AddToWaitGroup test passed")
}

// TestGo_WithPanicRecoveryAndWaitGroup tests Go() with panic recovery and wait group
func TestGo_WithPanicRecoveryAndWaitGroup(t *testing.T) {
	fmt.Println("\n=== TestGo_WithPanicRecoveryAndWaitGroup ===")
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

	// Spawn goroutines with panic recovery and wait group
	// Mix of panicking and normal goroutines
	var normalCompleted atomic.Int32

	// Normal goroutine
	err = localMgr.Go("mixed-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		normalCompleted.Add(1)
		return nil
	}, Local.WithPanicRecovery(true), Local.AddToWaitGroup("mixed-worker"))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Panicking goroutine (should be recovered)
	err = localMgr.Go("mixed-worker", func(ctx context.Context) error {
		panic("test panic")
	}, Local.WithPanicRecovery(true), Local.AddToWaitGroup("mixed-worker"))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for all to complete
	time.Sleep(100 * time.Millisecond)
	err = localMgr.WaitForFunction("mixed-worker")
	if err != nil {
		t.Fatalf("WaitForFunction() failed: %v", err)
	}

	if normalCompleted.Load() != 1 {
		t.Errorf("Expected 1 normal goroutine to complete, got %d", normalCompleted.Load())
	}

	fmt.Println("✓ WithPanicRecovery + AddToWaitGroup test passed")
}

// TestGo_AllOptions tests Go() with all three options
func TestGo_AllOptions(t *testing.T) {
	fmt.Println("\n=== TestGo_AllOptions ===")
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

	// Spawn goroutine with all options
	var completed atomic.Bool
	err = localMgr.Go("all-options-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		completed.Store(true)
		return nil
	}, Local.WithTimeout(200*time.Millisecond), Local.WithPanicRecovery(true), Local.AddToWaitGroup("all-options-worker"))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for completion
	err = localMgr.WaitForFunction("all-options-worker")
	if err != nil {
		t.Fatalf("WaitForFunction() failed: %v", err)
	}

	if !completed.Load() {
		t.Error("Goroutine should have completed")
	}

	fmt.Println("✓ All options test passed")
}

// TestGo_InvalidOption tests Go() with invalid option type (should be ignored gracefully)
func TestGo_InvalidOption(t *testing.T) {
	fmt.Println("\n=== TestGo_InvalidOption ===")
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

	// Spawn goroutine with invalid option (should be ignored)
	var completed atomic.Bool
	err = localMgr.Go("invalid-option-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		completed.Store(true)
		return nil
	}, "invalid-option") // Invalid option type

	if err != nil {
		t.Fatalf("Go() should not fail with invalid option: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if !completed.Load() {
		t.Error("Goroutine should have completed despite invalid option")
	}

	fmt.Println("✓ Invalid option test passed")
}

// TestGo_MultipleOptionsOrder tests that options can be applied in any order
func TestGo_MultipleOptionsOrder(t *testing.T) {
	fmt.Println("\n=== TestGo_MultipleOptionsOrder ===")
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

	// Test different order of options
	var completed atomic.Bool
	err = localMgr.Go("order-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		completed.Store(true)
		return nil
	}, Local.AddToWaitGroup("order-worker"), Local.WithTimeout(200*time.Millisecond), Local.WithPanicRecovery(true))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for completion
	err = localMgr.WaitForFunction("order-worker")
	if err != nil {
		t.Fatalf("WaitForFunction() failed: %v", err)
	}

	if !completed.Load() {
		t.Error("Goroutine should have completed")
	}

	fmt.Println("✓ Multiple options order test passed")
}

// TestGo_ContextCancellation tests that context is properly cancelled on timeout
func TestGo_ContextCancellation(t *testing.T) {
	fmt.Println("\n=== TestGo_ContextCancellation ===")
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

	// Test context cancellation on timeout
	var ctxCancelled atomic.Bool
	err = localMgr.Go("ctx-cancel-worker", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			ctxCancelled.Store(true)
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			return nil
		}
	}, Local.WithTimeout(100*time.Millisecond))

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	if !ctxCancelled.Load() {
		t.Error("Context should have been cancelled on timeout")
	}

	fmt.Println("✓ Context cancellation test passed")
}

// TestGo_EmptyOptions tests Go() with empty options slice
func TestGo_EmptyOptions(t *testing.T) {
	fmt.Println("\n=== TestGo_EmptyOptions ===")
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

	// Spawn goroutine with empty options (variadic)
	var completed atomic.Bool
	err = localMgr.Go("empty-opts-worker", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		completed.Store(true)
		return nil
	}) // No options provided

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if !completed.Load() {
		t.Error("Goroutine should have completed")
	}

	fmt.Println("✓ Empty options test passed")
}
