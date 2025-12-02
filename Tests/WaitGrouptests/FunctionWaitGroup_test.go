package WaitGrouptests

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


func TestFunctionWaitGroup_CreateAndRetrieve(t *testing.T) {
	fmt.Println("\n=== TestFunctionWaitGroup_CreateAndRetrieve ===")
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

	// Create wait group for a function
	fmt.Println("Creating wait group for function 'worker'...")
	wg1, err := localMgr.NewFunctionWaitGroup(context.Background(), "worker")
	if err != nil {
		t.Fatalf("NewFunctionWaitGroup() failed: %v", err)
	}
	fmt.Println("✓ Wait group created")

	// Retrieve the same wait group
	fmt.Println("Retrieving the same wait group...")
	wg2, err := localMgr.NewFunctionWaitGroup(context.Background(), "worker")
	if err != nil {
		t.Fatalf("Second NewFunctionWaitGroup() failed: %v", err)
	}

	// Should be the same instance
	if wg1 != wg2 {
		t.Error("Expected same wait group instance")
	} else {
		fmt.Println("✓ Same wait group instance returned")
	}
}

func TestFunctionWaitGroup_AutoManagement(t *testing.T) {
	fmt.Println("\n=== TestFunctionWaitGroup_AutoManagement ===")
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

	// Spawn 10 goroutines with same function name
	fmt.Println("Spawning 10 goroutines with function name 'worker'...")
	var counter atomic.Int32
	for i := 0; i < 10; i++ {
		err = localMgr.Go("worker", func(ctx context.Context) error {
			counter.Add(1)
			time.Sleep(100 * time.Millisecond)
			return nil
		}, Local.AddToWaitGroup("worker"))
		if err != nil {
			t.Fatalf("GoWithWaitGroup() failed: %v", err)
		}
	}
	fmt.Println("✓ All goroutines spawned")

	// Verify count
	count := localMgr.GetFunctionGoroutineCount("worker")
	fmt.Printf("Function goroutine count: %d\n", count)
	if count != 10 {
		t.Errorf("Expected 10 goroutines, got %d", count)
	}

	// Wait for all to complete
	fmt.Println("Waiting for all worker goroutines to complete...")
	err = localMgr.WaitForFunction("worker")
	if err != nil {
		t.Fatalf("WaitForFunction() failed: %v", err)
	}
	fmt.Println("✓ All goroutines completed")

	// Verify all executed
	if counter.Load() != 10 {
		t.Errorf("Expected counter to be 10, got %d", counter.Load())
	} else {
		fmt.Println("✓ All 10 goroutines executed")
	}
}

func TestFunctionWaitGroup_MultipleFunctions(t *testing.T) {
	fmt.Println("\n=== TestFunctionWaitGroup_MultipleFunctions ===")
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

	// Spawn goroutines with 3 different function names
	fmt.Println("Spawning goroutines with 3 different functions...")

	var counterA, counterB, counterC atomic.Int32

	// Function A - 5 goroutines (fast)
	for i := 0; i < 5; i++ {
		localMgr.Go("functionA", func(ctx context.Context) error {
			counterA.Add(1)
			time.Sleep(50 * time.Millisecond)
			return nil
		}, Local.AddToWaitGroup("functionA"))
	}
	fmt.Println("  ✓ FunctionA: 5 goroutines spawned")

	// Function B - 3 goroutines (medium)
	for i := 0; i < 3; i++ {
		localMgr.Go("functionB", func(ctx context.Context) error {
			counterB.Add(1)
			time.Sleep(100 * time.Millisecond)
			return nil
		}, Local.AddToWaitGroup("functionB"))
	}
	fmt.Println("  ✓ FunctionB: 3 goroutines spawned")

	// Function C - 2 goroutines (slow)
	for i := 0; i < 2; i++ {
		localMgr.Go("functionC", func(ctx context.Context) error {
			counterC.Add(1)
			time.Sleep(150 * time.Millisecond)
			return nil
		}, Local.AddToWaitGroup("functionC"))
	}
	fmt.Println("  ✓ FunctionC: 2 goroutines spawned")

	// Verify counts
	fmt.Println("\nVerifying goroutine counts per function...")
	countA := localMgr.GetFunctionGoroutineCount("functionA")
	countB := localMgr.GetFunctionGoroutineCount("functionB")
	countC := localMgr.GetFunctionGoroutineCount("functionC")

	fmt.Printf("  FunctionA: %d goroutines\n", countA)
	fmt.Printf("  FunctionB: %d goroutines\n", countB)
	fmt.Printf("  FunctionC: %d goroutines\n", countC)

	if countA != 5 || countB != 3 || countC != 2 {
		t.Errorf("Expected counts 5,3,2 got %d,%d,%d", countA, countB, countC)
	}

	// Wait for specific function
	fmt.Println("\nWaiting for functionA to complete...")
	err = localMgr.WaitForFunction("functionA")
	if err != nil {
		t.Fatalf("WaitForFunction(functionA) failed: %v", err)
	}
	fmt.Printf("✓ FunctionA completed (counter=%d)\n", counterA.Load())

	// Wait for all
	fmt.Println("Waiting for all functions to complete...")
	localMgr.WaitForFunction("functionB")
	localMgr.WaitForFunction("functionC")

	fmt.Printf("✓ All functions completed (A=%d, B=%d, C=%d)\n",
		counterA.Load(), counterB.Load(), counterC.Load())
}

func TestFunctionWaitGroup_FanOut(t *testing.T) {
	fmt.Println("\n=== TestFunctionWaitGroup_FanOut ===")
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

	// Fan-out pattern: spawn 20 workers to process items
	fmt.Println("Fan-out: Spawning 20 worker goroutines...")

	var processed atomic.Int32
	workItems := 100
	itemsPerWorker := workItems / 20

	for i := 0; i < 20; i++ {
		workerID := i
		err = localMgr.Go("worker", func(ctx context.Context) error {
			// Each worker processes some items
			for j := 0; j < itemsPerWorker; j++ {
				processed.Add(1)
				time.Sleep(2 * time.Millisecond) // Simulate work
			}
			fmt.Printf("  Worker %2d completed %d items\n", workerID, itemsPerWorker)
			return nil
		}, Local.AddToWaitGroup("worker"))
		if err != nil {
			t.Fatalf("GoWithWaitGroup() failed: %v", err)
		}
	}
	fmt.Println("✓ All workers spawned")

	// Wait for all workers to complete
	fmt.Println("\nWaiting for all workers to complete...")
	startTime := time.Now()
	err = localMgr.WaitForFunction("worker")
	if err != nil {
		t.Fatalf("WaitForFunction() failed: %v", err)
	}
	elapsed := time.Since(startTime)

	fmt.Printf("✓ All workers completed in %v\n", elapsed)
	fmt.Printf("Total items processed: %d/%d\n", processed.Load(), workItems)

	if processed.Load() != int32(workItems) {
		t.Errorf("Expected %d items processed, got %d", workItems, processed.Load())
	}
}

func TestFunctionWaitGroup_SelectiveShutdown(t *testing.T) {
	fmt.Println("\n=== TestFunctionWaitGroup_SelectiveShutdown ===")
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

	// Spawn long-running goroutines for 3 functions
	fmt.Println("Spawning long-running goroutines for functions A, B, C...")

	var counterA, counterB, counterC atomic.Int32

	// Function A - will be shutdown
	for i := 0; i < 3; i++ {
		localMgr.Go("functionA", func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					counterA.Add(1)
					return ctx.Err()
				case <-time.After(10 * time.Millisecond):
					// Keep running
				}
			}
		}, Local.AddToWaitGroup("functionA"))
	}

	// Function B - will be shutdown
	for i := 0; i < 2; i++ {
		localMgr.Go("functionB", func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					counterB.Add(1)
					return ctx.Err()
				case <-time.After(10 * time.Millisecond):
					// Keep running
				}
			}
		}, Local.AddToWaitGroup("functionB"))
	}

	// Function C - will keep running
	for i := 0; i < 2; i++ {
		localMgr.Go("functionC", func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					counterC.Add(1)
					return ctx.Err()
				case <-time.After(10 * time.Millisecond):
					// Keep running
				}
			}
		}, Local.AddToWaitGroup("functionC"))
	}

	fmt.Println("✓ All goroutines spawned")
	time.Sleep(50 * time.Millisecond) // Let them run

	// Shutdown only functionB
	fmt.Println("\nShutting down functionB...")
	err = localMgr.ShutdownFunction("functionB", 500*time.Millisecond)
	if err != nil {
		t.Fatalf("ShutdownFunction(functionB) failed: %v", err)
	}
	fmt.Printf("✓ FunctionB shutdown (cancelled %d goroutines)\n", counterB.Load())

	// Verify B is shutdown but A and C still running
	if counterB.Load() != 2 {
		t.Errorf("Expected 2 functionB goroutines cancelled, got %d", counterB.Load())
	}

	countA := localMgr.GetFunctionGoroutineCount("functionA")
	countC := localMgr.GetFunctionGoroutineCount("functionC")

	fmt.Printf("FunctionA still has %d goroutines running\n", countA)
	fmt.Printf("FunctionC still has %d goroutines running\n", countC)

	if countA != 3 || countC != 2 {
		t.Errorf("Expected A=3, C=2 still running, got A=%d, C=%d", countA, countC)
	}

	// Cleanup
	localMgr.ShutdownFunction("functionA", 500*time.Millisecond)
	localMgr.ShutdownFunction("functionC", 500*time.Millisecond)

	fmt.Println("✓ Selective shutdown successful")
}

func TestFunctionWaitGroup_Timeout(t *testing.T) {
	fmt.Println("\n=== TestFunctionWaitGroup_Timeout ===")
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

	// Spawn long-running goroutine
	fmt.Println("Spawning long-running goroutine...")
	localMgr.Go("slowWorker", func(ctx context.Context) error {
		time.Sleep(2 * time.Second)
		return nil
	}, Local.AddToWaitGroup("slowWorker"))

	// Wait with short timeout
	fmt.Println("Waiting with 100ms timeout (should timeout)...")
	completed := localMgr.WaitForFunctionWithTimeout("slowWorker", 100*time.Millisecond)

	if completed {
		t.Error("Expected timeout, but wait completed")
	} else {
		fmt.Println("✓ Timeout occurred as expected")
	}

	// Wait with long timeout
	fmt.Println("Waiting with 3s timeout (should complete)...")
	completed = localMgr.WaitForFunctionWithTimeout("slowWorker", 3*time.Second)

	if !completed {
		t.Error("Expected completion, but timeout occurred")
	} else {
		fmt.Println("✓ Completed within timeout")
	}
}
