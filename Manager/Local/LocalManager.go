package Local

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
	"github.com/neerajchowdary889/GoRoutinesManager/metrics"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Errors"
)

type LocalManagerStruct struct {
	AppName   string
	LocalName string
}

func NewLocalManager(appName, localName string) Interface.LocalGoroutineManagerInterface {
	return &LocalManagerStruct{
		AppName:   appName,
		LocalName: localName,
	}
}

func (LM *LocalManagerStruct) CreateLocal(localName string) (*types.LocalManager, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metrics.RecordManagerOperationDuration("local", "create", duration, LM.AppName)
	}()

	// First get the app manager
	appManager, err := types.GetAppManager(LM.AppName)
	if err != nil {
		metrics.RecordOperationError("manager", "create_local", "get_app_manager_failed")
		return nil, err
	}

	// Directly call the CreateLocal method of the app manager
	// CreateLocal function will handle the checking and creation of the local manager
	localManager, err := appManager.CreateLocal(localName)
	switch err {
	case Errors.ErrLocalManagerNotFound:
		metrics.RecordOperationError("manager", "create_local", "local_manager_not_found")
		return nil, fmt.Errorf("%w: %s", Errors.ErrLocalManagerNotFound, localName)
	case Errors.WrngLocalManagerAlreadyExists:
		// Return the existing local manager and also return error as nil
		return localManager, nil
	default:
		// Fill the structs
		localManager.SetLocalContext().
			SetLocalMutex().
			SetLocalWaitGroup()
		// Record operation
		metrics.RecordManagerOperation("local", "create", LM.AppName)
	}
	return localManager, nil
}

// Shutdowner
func (LM *LocalManagerStruct) Shutdown(safe bool) error {
	startTime := time.Now()
	shutdownType := "unsafe"
	if safe {
		shutdownType = "safe"
	}

	defer func() {
		duration := time.Since(startTime)
		metrics.RecordShutdownDuration("local", shutdownType, duration, LM.AppName, LM.LocalName)
	}()

	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		metrics.RecordOperationError("manager", "shutdown", "get_local_manager_failed")
		return err
	}

	// Record shutdown operation
	metrics.RecordManagerOperation("local", "shutdown", LM.AppName)

	// Track all function names for cleanup
	var functionNames map[string]bool
	var routines []*types.Routine

	// Defer cleanup to ensure it happens even on panic or early return
	defer func() {
		// Clean up all function wait groups
		// Safe to range over nil map - it just does nothing
		for functionName := range functionNames {
			localManager.RemoveFunctionWg(functionName)
		}
	}()

	// Panic recovery to ensure cleanup happens
	defer func() {
		if r := recover(); r != nil {
			// Ensure cleanup happens even on panic
			// Safe to range over nil map - it just does nothing
			for functionName := range functionNames {
				localManager.RemoveFunctionWg(functionName)
			}
			// Re-panic after cleanup
			panic(r)
		}
	}()

	if safe {
		// Safe shutdown: try graceful shutdown first, then force cancel hanging goroutines

		// Step 1: Get all routines and function names
		var err error
		routines, err = LM.GetAllGoroutines()
		if err != nil {
			metrics.RecordOperationError("manager", "shutdown", "get_goroutines_failed")
			return err
		}

		functionNames = make(map[string]bool)
		for _, routine := range routines {
			functionNames[routine.GetFunctionName()] = true
		}

		// Step 2: Try to shutdown each function gracefully with timeout
		shutdownTimeout := types.ShutdownTimeout
		for functionName := range functionNames {
			// Try graceful shutdown with timeout
			_ = LM.ShutdownFunction(functionName, shutdownTimeout)
			// Note: ShutdownFunction handles cleanup on success, but we'll clean up all in defer
		}

		// Step 3: Wait for main wait group with timeout
		done := make(chan struct{})
		go func() {
			// Lock to safely read Wg pointer to avoid race condition
			localManager.LockAppReadMutex()
			wg := localManager.Wg
			localManager.UnlockAppReadMutex()
			if wg != nil {
				wg.Wait()
			}
			close(done)
		}()

		// Wait with timeout
		select {
		case <-done:
			// All goroutines completed gracefully
			// Cleanup will happen in defer
			return nil
		case <-time.After(shutdownTimeout):
			// Timeout - some goroutines are still hanging
			// Fall through to force cancel
		}

		// Step 4: Force cancel any remaining hanging goroutines
		remainingRoutines, err := LM.GetAllGoroutines()
		if err == nil {
			// Record remaining goroutines after timeout
			metrics.RecordShutdownGoroutinesRemaining("local", LM.AppName, LM.LocalName, len(remainingRoutines))
			for _, routine := range remainingRoutines {
				cancel := routine.GetCancel()
				if cancel != nil {
					cancel()
				}
				// Remove routine from map to prevent memory leak
				localManager.RemoveRoutine(routine, false)
			}
		}

		// Cancel the local manager's context
		if localManager.Cancel != nil {
			localManager.Cancel()
		}

	} else {
		// Unsafe shutdown: cancel all contexts immediately
		// Get all routines and cancel their contexts
		var err error
		routines, err = LM.GetAllGoroutines()
		if err != nil {
			return err
		}

		// Track function names for cleanup
		functionNames = make(map[string]bool)
		for _, routine := range routines {
			functionNames[routine.GetFunctionName()] = true
		}

		// Cancel all routine contexts and remove from map
		for _, routine := range routines {
			cancel := routine.GetCancel()
			if cancel != nil {
				cancel()
			}
			// Remove routine from map to prevent memory leak
			localManager.RemoveRoutine(routine, false)
		}

		// Cancel the local manager's context
		if localManager.Cancel != nil {
			localManager.Cancel()
		}
	}

	return nil
}

// FunctionShutdowner
func (LM *LocalManagerStruct) ShutdownFunction(functionName string, timeout time.Duration) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metrics.RecordGoroutineOperationDuration("shutdown_function", duration, LM.AppName, LM.LocalName, functionName)
	}()

	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		metrics.RecordOperationError("function", "shutdown", "get_local_manager_failed")
		return err
	}

	// Record operation
	metrics.RecordFunctionOperation("shutdown", LM.AppName, LM.LocalName, functionName)

	// Get all routines
	routines, err := LM.GetAllGoroutines()
	if err != nil {
		metrics.RecordOperationError("function", "shutdown", "get_goroutines_failed")
		return err
	}

	// Track routines for this function for cleanup
	var functionRoutines []*types.Routine

	// Cancel all routines with this function name
	for _, routine := range routines {
		if routine.GetFunctionName() == functionName {
			functionRoutines = append(functionRoutines, routine)
			cancel := routine.GetCancel()
			if cancel != nil {
				cancel()
			}
		}
	}

	// Wait for completion with timeout
	completed := LM.WaitForFunctionWithTimeout(functionName, timeout)
	if !completed {
		// Timeout occurred - clean up routines and wait group
		for _, routine := range functionRoutines {
			// Remove routine from map to prevent memory leak
			localManager.RemoveRoutine(routine, false)
		}
		// Clean up the wait group even on timeout
		localManager.RemoveFunctionWg(functionName)
		return fmt.Errorf("shutdown timeout for function: %s", functionName)
	}

	// Clean up the wait group on success
	localManager.RemoveFunctionWg(functionName)

	return nil
}

// GoroutineSpawner
// Go spawns a new goroutine, tracks it in the LocalManager, and returns the routine ID.
// The goroutine is spawned with a context derived from the LocalManager's parent context.
// The done channel is closed when the goroutine completes.
//
// Options can be provided to configure the goroutine:
//   - WithTimeout(duration): Sets a timeout for the goroutine. Context is cancelled on timeout.
//   - WithPanicRecovery(enabled): Enables panic recovery. Panics are logged and goroutine completes normally.
//   - AddToWaitGroup(functionName): Adds the goroutine to a function wait group for coordinated shutdown.
//
// Example:
//
//	localMgr.Go("worker", func(ctx context.Context) error { ... },
//	    WithTimeout(5*time.Second),
//	    WithPanicRecovery(true),
//	    AddToWaitGroup("worker"))
func (LM *LocalManagerStruct) Go(functionName string, workerFunc func(ctx context.Context) error, opts ...Interface.GoroutineOption) error {
	// Apply default options
	options := defaultGoroutineOptions()
	for _, opt := range opts {
		// Type assert to Option (defined in this package)
		if localOpt, ok := opt.(Option); ok {
			localOpt(options)
		}
	}
	return LM.spawnGoroutine(functionName, workerFunc, options)
}

// spawnGoroutine is the internal implementation for spawning goroutines.
// It accepts options to configure timeout, panic recovery, and wait group behavior.
func (LM *LocalManagerStruct) spawnGoroutine(functionName string, workerFunc func(ctx context.Context) error, opts *goroutineOptions) error {
	// Get the types.LocalManager instance
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return err
	}

	var wg *sync.WaitGroup
	if opts.waitGroupName != "" {
		// Get or create function wait group using the specified function name
		wg, err = LM.NewFunctionWaitGroup(context.Background(), opts.waitGroupName)
		if err != nil {
			return err
		}
		// Increment wait group BEFORE spawning goroutine
		wg.Add(1)
	}

	// Always add to LocalManager's main wait group for safe shutdown
	if localManager.Wg != nil {
		localManager.Wg.Add(1)
	}

	// Create a child context with cancel for this routine
	routineCtx, cancel := localManager.SpawnChild()

	// Apply timeout if specified
	var timeoutCancel context.CancelFunc
	if opts.timeout != nil {
		routineCtx, timeoutCancel = context.WithTimeout(routineCtx, *opts.timeout)
		// Combine cancellations: when timeout expires or explicit cancel is called
		originalCancel := cancel
		cancel = func() {
			originalCancel()
			if timeoutCancel != nil {
				timeoutCancel()
			}
		}
	}

	// Create the done channel (bidirectional, buffered size 1)
	// This allows non-blocking close even if nothing is reading
	doneChan := make(chan struct{}, 1)

	// Create a new Routine instance
	routine := localManager.NewGoRoutine(functionName).
		SetContext(routineCtx).
		SetCancel(cancel).
		SetDone(doneChan) // Override the channel created in NewGoRoutine

	// Record goroutine creation and measure creation duration
	createStartTime := time.Now()
	metrics.RecordGoroutineOperation("create", LM.AppName, LM.LocalName, functionName)

	// Spawn the goroutine
	go func() {
		startTimeNano := time.Now().UnixNano()
		defer func() {
			// Handle panic recovery (enabled by default for production safety)
			if opts.panicRecovery {
				if r := recover(); r != nil {
					// Log panic details via metrics
					metrics.RecordOperationError("goroutine", "panic", fmt.Sprintf("function: %s, panic: %v", functionName, r))
					// Panic is recovered, continue with normal cleanup
				}
			}

			// Record goroutine completion
			metrics.RecordGoroutineCompletion(LM.AppName, LM.LocalName, functionName, startTimeNano)
			metrics.RecordGoroutineOperation("complete", LM.AppName, LM.LocalName, functionName)

			if opts.waitGroupName != "" && wg != nil {
				// Decrement function wait group when routine completes
				wg.Done()
			}
			// Always decrement LocalManager's main wait group
			if localManager.Wg != nil {
				localManager.Wg.Done()
			}
			// Close the done channel when routine completes
			// The done channel is buffered (size 1) so this won't block
			close(doneChan)

			// Explicitly cancel the routine's context to ensure proper cleanup
			// This ensures any resources tied to the context are released immediately
			// Context inheritance handles parent cancellation, but explicit cleanup is better
			if cancel != nil {
				cancel()
			}

			// CRITICAL: Remove routine from tracking map to prevent memory leak
			// This must be done after all cleanup to ensure the routine is fully completed
			// Using safe=false since the routine is already completing naturally
			// Note: RemoveRoutine also cancels the context, but we've already done it above
			// for explicit cleanup. RemoveRoutine's cancel is idempotent (safe to call twice).
			localManager.RemoveRoutine(routine, false)
		}()

		// Execute the worker function with the routine's context
		// Panics will be caught and recovered by the defer block above (enabled by default)
		_ = workerFunc(routineCtx)
	}()

	// Record creation operation duration (time to spawn goroutine, should be very fast)
	createDuration := time.Since(createStartTime)
	metrics.RecordGoroutineOperationDuration("create", createDuration, LM.AppName, LM.LocalName, functionName)

	return nil
}

// GoroutineLister
func (LM *LocalManagerStruct) GetAllGoroutines() ([]*types.Routine, error) {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return nil, err
	}

	routines := localManager.GetRoutines()
	// Convert map to slice
	result := make([]*types.Routine, 0, len(routines))
	for _, routine := range routines {
		result = append(result, routine)
	}
	return result, nil
}

func (LM *LocalManagerStruct) GetGoroutineCount() int {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return 0
	}
	return localManager.GetRoutineCount()
}

// FunctionWaitGroupCreator
func (LM *LocalManagerStruct) NewFunctionWaitGroup(ctx context.Context, functionName string) (*sync.WaitGroup, error) {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		metrics.RecordOperationError("function", "wait_group_create", "get_local_manager_failed")
		return nil, err
	}

	// Check if wait group already exists
	wg, err := localManager.GetFunctionWg(functionName)
	if err == nil {
		// Already exists, return it
		return wg, nil
	}

	// Create new wait group
	localManager.AddFunctionWg(functionName)

	// Record operation
	metrics.RecordFunctionOperation("wait_group_create", LM.AppName, LM.LocalName, functionName)

	// Retrieve and return the newly created wait group
	return localManager.GetFunctionWg(functionName)
}

// Multiple go routines can have same function name
// This function is to get the routines by function name
func (LM *LocalManagerStruct) GetRoutinesByFunctionName(functionName string) ([]*types.Routine, error) {
	result := make([]*types.Routine, 0)
	routines, err := LM.GetAllGoroutines()
	if err != nil {
		return nil, err
	}
	for _, routine := range routines {
		if routine.GetFunctionName() == functionName {
			result = append(result, routine)
		}
	}
	return result, nil	
}
