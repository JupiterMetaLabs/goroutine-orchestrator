package manager_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/app"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/local"
	common "github.com/JupiterMetaLabs/goroutine-orchestrator/test/common"
)

func TestLocalManager_CreateLocal(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_CreateLocal ===")
	common.ResetGlobalState()

	// Setup: create app first
	fmt.Println("Creating app manager...")
	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}
	fmt.Println("✓ App manager created")

	// Create local manager
	fmt.Println("Creating local manager...")
	localMgr := local.NewLocalManager("test-app", "test-local")
	local, err := localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	if local == nil {
		t.Fatal("CreateLocal() returned nil")
	}
	fmt.Printf("✓ Local manager created: %s\n", local.GetLocalName())
}

func TestLocalManager_CreateLocal_Idempotent(t *testing.T) {
	common.ResetGlobalState()

	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")

	// Create first time
	local1, err := localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("First CreateLocal() failed: %v", err)
	}

	// Create second time (should return existing)
	local2, err := localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("Second CreateLocal() failed: %v", err)
	}

	// Should be the same instance
	if local1 != local2 {
		t.Error("CreateLocal() should return same instance on subsequent calls")
	}
}

func TestLocalManager_Go_SpawnGoroutine(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_Go_SpawnGoroutine ===")
	common.ResetGlobalState()

	// Setup
	fmt.Println("Setting up app and local managers...")
	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}
	fmt.Println("✓ Setup complete")

	// Spawn a goroutine
	fmt.Println("Spawning goroutine with Go()...")
	executed := atomic.Bool{}
	err = localMgr.Go("testFunc", func(ctx context.Context) error {
		fmt.Println("  → Goroutine executing...")
		executed.Store(true)
		fmt.Println("  → Goroutine completed")
		return nil
	})

	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}
	fmt.Println("✓ Goroutine spawned")

	// Wait a bit for goroutine to execute
	fmt.Println("Waiting for goroutine to execute...")
	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("Goroutine was not executed")
	} else {
		fmt.Println("✓ Goroutine executed successfully")
	}

	// Verify routine was tracked
	count := localMgr.GetGoroutineCount()
	fmt.Printf("Goroutine count: %d\n", count)
	if count != 1 {
		t.Errorf("Expected 1 tracked goroutine, got %d", count)
	} else {
		fmt.Println("✓ Goroutine tracked correctly")
	}
}

func TestLocalManager_Go_MultipleGoroutines(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_Go_MultipleGoroutines ===")
	common.ResetGlobalState()

	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Spawn multiple goroutines
	fmt.Println("Spawning 5 goroutines...")
	for i := 0; i < 5; i++ {
		err = localMgr.Go("testFunc", func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Fatalf("Go() failed: %v", err)
		}
		fmt.Printf("  Spawned goroutine %d/5\n", i+1)
	}
	fmt.Println("✓ All goroutines spawned")

	// Verify all were tracked
	count := localMgr.GetGoroutineCount()
	fmt.Printf("Total goroutines tracked: %d\n", count)
	if count != 5 {
		t.Errorf("Expected 5 tracked goroutines, got %d", count)
	} else {
		fmt.Println("✓ All goroutines tracked correctly")
	}
}

func TestLocalManager_GetAllGoroutines(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_GetAllGoroutines ===")
	common.ResetGlobalState()

	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Initially empty
	fmt.Println("Checking initial state (should be empty)...")
	routines, err := localMgr.GetAllGoroutines()
	if err != nil {
		t.Fatalf("GetAllGoroutines() failed: %v", err)
	}
	if len(routines) != 0 {
		t.Errorf("Expected 0 goroutines, got %d", len(routines))
	} else {
		fmt.Println("✓ Initial state is empty")
	}

	// Spawn some goroutines
	fmt.Println("Spawning 3 goroutines...")
	for i := 0; i < 3; i++ {
		localMgr.Go("testFunc", func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}
	fmt.Println("✓ Goroutines spawned")

	// Get all
	fmt.Println("Retrieving all goroutines...")
	routines, err = localMgr.GetAllGoroutines()
	if err != nil {
		t.Fatalf("GetAllGoroutines() failed: %v", err)
	}
	fmt.Printf("Found %d goroutines\n", len(routines))
	if len(routines) != 3 {
		t.Errorf("Expected 3 goroutines, got %d", len(routines))
	}

	// Verify each routine has required fields
	fmt.Println("Verifying goroutine metadata...")
	for i, routine := range routines {
		fmt.Printf("  Goroutine %d:\n", i+1)
		fmt.Printf("    ID: %s\n", routine.GetID())
		fmt.Printf("    Function: %s\n", routine.GetFunctionName())
		fmt.Printf("    Started At: %d\n", routine.GetStartedAt())
		if routine.GetID() == "" {
			t.Error("Routine should have an ID")
		}
		if routine.GetFunctionName() == "" {
			t.Error("Routine should have a function name")
		}
		if routine.GetStartedAt() == 0 {
			t.Error("Routine should have a start timestamp")
		}
	}
	fmt.Println("✓ All metadata verified")
}

func TestLocalManager_GetGoroutineCount(t *testing.T) {
	common.ResetGlobalState()

	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Initially 0
	count := localMgr.GetGoroutineCount()
	if count != 0 {
		t.Errorf("Expected 0 goroutines, got %d", count)
	}

	// Add some
	localMgr.Go("func1", func(ctx context.Context) error { return nil })
	localMgr.Go("func2", func(ctx context.Context) error { return nil })

	count = localMgr.GetGoroutineCount()
	if count != 2 {
		t.Errorf("Expected 2 goroutines, got %d", count)
	}
}

func TestLocalManager_ContextPropagation(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_ContextPropagation ===")
	common.ResetGlobalState()

	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Spawn a goroutine that checks context
	fmt.Println("Spawning goroutine to test context propagation...")
	contextReceived := atomic.Bool{}
	err = localMgr.Go("testFunc", func(ctx context.Context) error {
		if ctx != nil {
			fmt.Println("  → Context received in goroutine: ✓")
			contextReceived.Store(true)
		} else {
			fmt.Println("  → Context is nil: ✗")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for execution
	fmt.Println("Waiting for goroutine to execute...")
	time.Sleep(100 * time.Millisecond)

	if !contextReceived.Load() {
		t.Error("Goroutine should receive a valid context")
	} else {
		fmt.Println("✓ Context propagated successfully")
	}
}

func TestLocalManager_GoroutineCompletion(t *testing.T) {
	common.ResetGlobalState()

	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Spawn a short-lived goroutine
	completed := atomic.Bool{}
	err = localMgr.Go("testFunc", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		completed.Store(true)
		return nil
	})
	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for completion
	time.Sleep(150 * time.Millisecond)

	if !completed.Load() {
		t.Error("Goroutine should have completed")
	}
}

func TestLocalManager_MultipleLocalManagers(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_MultipleLocalManagers ===")
	common.ResetGlobalState()

	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	// Create multiple local managers
	fmt.Println("Creating 2 local managers...")
	local1Mgr := local.NewLocalManager("test-app", "local1")
	local2Mgr := local.NewLocalManager("test-app", "local2")

	_, err = local1Mgr.CreateLocal("local1")
	if err != nil {
		t.Fatalf("CreateLocal(local1) failed: %v", err)
	}
	fmt.Println("✓ Local1 created")

	_, err = local2Mgr.CreateLocal("local2")
	if err != nil {
		t.Fatalf("CreateLocal(local2) failed: %v", err)
	}
	fmt.Println("✓ Local2 created")

	// Spawn goroutines in each
	fmt.Println("Spawning goroutines in local1 (2) and local2 (1)...")
	local1Mgr.Go("func1", func(ctx context.Context) error { return nil })
	local1Mgr.Go("func2", func(ctx context.Context) error { return nil })
	local2Mgr.Go("func3", func(ctx context.Context) error { return nil })
	fmt.Println("✓ Goroutines spawned")

	// Verify counts are independent
	fmt.Println("Verifying independent counts...")
	count1 := local1Mgr.GetGoroutineCount()
	fmt.Printf("  Local1 goroutine count: %d\n", count1)
	if count1 != 2 {
		t.Errorf("Local1 should have 2 goroutines, got %d", count1)
	}

	count2 := local2Mgr.GetGoroutineCount()
	fmt.Printf("  Local2 goroutine count: %d\n", count2)
	if count2 != 1 {
		t.Errorf("Local2 should have 1 goroutine, got %d", count2)
	}
	fmt.Println("✓ Counts are independent and correct")
}

func TestLocalManager_ErrorHandling(t *testing.T) {
	common.ResetGlobalState()

	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Spawn a goroutine that returns an error
	errorReturned := atomic.Bool{}
	err = localMgr.Go("testFunc", func(ctx context.Context) error {
		errorReturned.Store(true)
		return context.Canceled
	})
	if err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	// Wait for execution
	time.Sleep(100 * time.Millisecond)

	if !errorReturned.Load() {
		t.Error("Goroutine should have executed and returned error")
	}
}


func TestLocalManager_ComplexOperationsWithArguments(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_ComplexOperationsWithArguments ===")
	common.ResetGlobalState()

	// Setup
	fmt.Println("Setting up app and local managers...")
	appMgr := app.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}
	fmt.Println("✓ Setup complete")

	// Shared state for results
	type Result struct {
		RoutineID   int
		Operation   string
		Input       int
		Output      int64
		ProcessTime time.Duration
	}

	var results []Result
	var resultsMutex sync.Mutex

	// Different operation functions
	addOperation := func(a, b int) int64 {
		time.Sleep(20 * time.Millisecond) // Simulate work
		return int64(a + b)
	}

	multiplyOperation := func(a, b int) int64 {
		time.Sleep(30 * time.Millisecond) // Simulate work
		return int64(a * b)
	}

	computeFactorial := func(n int) int64 {
		time.Sleep(25 * time.Millisecond) // Simulate work
		if n <= 1 {
			return 1
		}
		result := int64(1)
		for i := 2; i <= n; i++ {
			result *= int64(i)
		}
		return result
	}

	// Spawn goroutines with different operations
	fmt.Println("\nSpawning goroutines with different operations...")

	// Type 1: Addition operations (routines 1-5)
	for i := 1; i <= 5; i++ {
		routineID := i
		inputValue := i * 10

		err = localMgr.Go(fmt.Sprintf("add-operation-%d", routineID), func(ctx context.Context) error {
			startTime := time.Now()

			// Pass arguments via closure
			result := addOperation(inputValue, 100)

			elapsed := time.Since(startTime)
			fmt.Printf("  → ADD Routine %d: %d + 100 = %d (took %v)\n",
				routineID, inputValue, result, elapsed)

			// Store result
			resultsMutex.Lock()
			results = append(results, Result{
				RoutineID:   routineID,
				Operation:   "ADD",
				Input:       inputValue,
				Output:      result,
				ProcessTime: elapsed,
			})
			resultsMutex.Unlock()

			return nil
		})

		if err != nil {
			t.Fatalf("Go() failed for add routine %d: %v", i, err)
		}
	}

	// Type 2: Multiplication operations (routines 6-10)
	for i := 6; i <= 10; i++ {
		routineID := i
		inputValue := i - 5 // 1-5

		err = localMgr.Go(fmt.Sprintf("multiply-operation-%d", routineID), func(ctx context.Context) error {
			startTime := time.Now()

			// Pass arguments via closure
			result := multiplyOperation(inputValue, 7)

			elapsed := time.Since(startTime)
			fmt.Printf("  → MULTIPLY Routine %d: %d × 7 = %d (took %v)\n",
				routineID, inputValue, result, elapsed)

			// Store result
			resultsMutex.Lock()
			results = append(results, Result{
				RoutineID:   routineID,
				Operation:   "MULTIPLY",
				Input:       inputValue,
				Output:      result,
				ProcessTime: elapsed,
			})
			resultsMutex.Unlock()

			return nil
		})

		if err != nil {
			t.Fatalf("Go() failed for multiply routine %d: %v", i, err)
		}
	}

	// Type 3: Factorial operations (routines 11-15)
	for i := 11; i <= 15; i++ {
		routineID := i
		inputValue := i - 10 // 1-5

		err = localMgr.Go(fmt.Sprintf("factorial-operation-%d", routineID), func(ctx context.Context) error {
			startTime := time.Now()

			// Pass arguments via closure
			result := computeFactorial(inputValue)

			elapsed := time.Since(startTime)
			fmt.Printf("  → FACTORIAL Routine %d: %d! = %d (took %v)\n",
				routineID, inputValue, result, elapsed)

			// Store result
			resultsMutex.Lock()
			results = append(results, Result{
				RoutineID:   routineID,
				Operation:   "FACTORIAL",
				Input:       inputValue,
				Output:      result,
				ProcessTime: elapsed,
			})
			resultsMutex.Unlock()

			return nil
		})

		if err != nil {
			t.Fatalf("Go() failed for factorial routine %d: %v", i, err)
		}
	}

	fmt.Println("✓ All 15 goroutines spawned")

	// Verify tracking
	count := localMgr.GetGoroutineCount()
	fmt.Printf("\nTotal goroutines tracked: %d\n", count)
	if count != 15 {
		t.Errorf("Expected 15 goroutines, got %d", count)
	}

	// Wait for all to complete
	fmt.Println("\nWaiting for all operations to complete...")
	time.Sleep(1 * time.Second)

	// Analyze results
	fmt.Println("\n=== Results Summary ===")
	resultsMutex.Lock()
	defer resultsMutex.Unlock()

	if len(results) != 15 {
		t.Errorf("Expected 15 results, got %d", len(results))
	}

	// Group by operation type
	addResults := []Result{}
	multiplyResults := []Result{}
	factorialResults := []Result{}

	for _, r := range results {
		switch r.Operation {
		case "ADD":
			addResults = append(addResults, r)
		case "MULTIPLY":
			multiplyResults = append(multiplyResults, r)
		case "FACTORIAL":
			factorialResults = append(factorialResults, r)
		}
	}

	fmt.Printf("\nADD Operations: %d\n", len(addResults))
	for _, r := range addResults {
		fmt.Printf("  Routine %2d: Input=%d, Output=%d, Time=%v\n",
			r.RoutineID, r.Input, r.Output, r.ProcessTime)
	}

	fmt.Printf("\nMULTIPLY Operations: %d\n", len(multiplyResults))
	for _, r := range multiplyResults {
		fmt.Printf("  Routine %2d: Input=%d, Output=%d, Time=%v\n",
			r.RoutineID, r.Input, r.Output, r.ProcessTime)
	}

	fmt.Printf("\nFACTORIAL Operations: %d\n", len(factorialResults))
	for _, r := range factorialResults {
		fmt.Printf("  Routine %2d: Input=%d, Output=%d, Time=%v\n",
			r.RoutineID, r.Input, r.Output, r.ProcessTime)
	}

	// Verify some specific results
	fmt.Println("\n=== Verification ===")

	// Check factorial of 5 should be 120
	found := false
	for _, r := range factorialResults {
		if r.Input == 5 && r.Output == 120 {
			found = true
			fmt.Println("✓ Factorial(5) = 120 is correct")
			break
		}
	}
	if !found {
		t.Error("Expected to find factorial(5) = 120")
	}

	// Check multiply 5 * 7 should be 35
	found = false
	for _, r := range multiplyResults {
		if r.Input == 5 && r.Output == 35 {
			found = true
			fmt.Println("✓ Multiply(5, 7) = 35 is correct")
			break
		}
	}
	if !found {
		t.Error("Expected to find multiply(5, 7) = 35")
	}

	fmt.Println("\n✓ All complex operations completed successfully!")
}
