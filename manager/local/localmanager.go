package local

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/errors"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/interfaces"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/metrics"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/types"
)

// LocalManagerStruct manages goroutines for a specific module or file within an application.
// It provides methods to spawn, track, and shutdown goroutines with advanced features like
// function-level wait groups, panic recovery, timeouts, and hierarchical shutdown coordination.
// This is the third level in the manager hierarchy (Global → App → Local → Routine).
type LocalManagerStruct struct {
	// AppName is the name of the parent application manager
	AppName string
	// LocalName is the unique identifier for this local manager within the app
	LocalName string
}

// NewLocalManager creates and returns a new LocalManager instance.
// This constructor does not initialize the local manager - call CreateLocal() to register it
// with the app manager and set up tracking structures.
//
// Parameters:
//   - appName: Name of the parent application manager
//   - localName: Unique identifier for this local manager (e.g., "handlers", "workers")
//
// Returns:
//   - An implementation of LocalGoroutineManagerInterface
//
// Example:
//
//	localMgr := local.NewLocalManager("api-server", "http-handlers")
//	local, err := localMgr.CreateLocal("http-handlers")
func NewLocalManager(appName, localName string) interfaces.LocalGoroutineManagerInterface {
	return &LocalManagerStruct{
		AppName:   appName,
		LocalName: localName,
	}
}

// CreateLocal initializes and registers the local manager with its parent app manager.
// This method is idempotent - calling it multiple times returns the existing local manager.
//
// Parameters:
//   - localName: The name for this local manager (must match the name used in constructor)
//
// The method performs the following operations:
//   - Retrieves the parent app manager
//   - Checks if local manager already exists (returns existing if found)
//   - Creates local-level context derived from app context
//   - Initializes local mutex for thread-safe operations
//   - Sets up local wait group for coordinated shutdown
//   - Records metrics for the create operation
//
// Returns:
//   - *types.LocalManager: The initialized local manager instance
//   - error: nil on success, error if app manager not found or creation fails
//
// Example:
//
//	localMgr := local.NewLocalManager("api-server", "handlers")
//	local, err := localMgr.CreateLocal("handlers")
//	if err != nil {
//	    log.Fatalf("Failed to create local manager: %v", err)
//	}
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
	case errors.ErrLocalManagerNotFound:
		metrics.RecordOperationError("manager", "create_local", "local_manager_not_found")
		return nil, fmt.Errorf("%w: %s", errors.ErrLocalManagerNotFound, localName)
	case errors.WrngLocalManagerAlreadyExists:
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

// Shutdown gracefully or forcefully shuts down all goroutines managed by this local manager.
// This method implements a sophisticated shutdown strategy with graceful → timeout → force cancellation.
//
// Parameters:
//   - safe: If true, performs graceful shutdown with timeout protection.
//     If false, immediately cancels all goroutines without waiting.
//
// Safe Shutdown Flow (safe=true):
//  1. Collects all tracked goroutines and their function names
//  2. Attempts graceful shutdown per function with timeout
//  3. Waits for main wait group with global shutdown timeout
//  4. If timeout occurs: force cancels remaining goroutines
//  5. Cleans up all function wait groups (via defer)
//  6. Removes all routines from tracking map
//  7. Cancels local manager's context
//
// Unsafe Shutdown Flow (safe=false):
//  1. Immediately cancels all routine contexts
//  2. Removes all routines from tracking map
//  3. Cancels local manager's context
//  4. Cleans up all function wait groups
//  5. No waiting for goroutines to complete
//
// Returns:
//   - error: nil on success, error if local manager not found
//
// Note: Function wait group cleanup is guaranteed via defer, even on panic or timeout.
//
// Example:
//
//	// Graceful shutdown with timeout protection
//	if err := localMgr.Shutdown(true); err != nil {
//	    log.Printf("Shutdown error: %v", err)
//	}
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
			wg := localManager.GetLocalWaitGroup()
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

// ShutdownFunction gracefully shuts down all goroutines with a specific function name.
// This allows selective shutdown of goroutine groups without affecting others.
//
// Parameters:
//   - functionName: Name of the function to shutdown (all goroutines with this name)
//   - timeout: Maximum time to wait for graceful shutdown before force cancellation
//
// Shutdown Process:
//  1. Retrieves all goroutines with the specified function name
//  2. Cancels their contexts to signal shutdown
//  3. Waits for function wait group with timeout
//  4. If timeout: removes routines from tracking and cleans up wait group
//  5. If success: cleans up wait group
//  6. Records metrics for shutdown duration
//
// Returns:
//   - error: nil if all goroutines shutdown within timeout
//     Returns error with timeout message if timeout occurs
//
// Example:
//
//	// Shutdown all "worker" goroutines with 10 second timeout
//	err := localMgr.ShutdownFunction("worker", 10*time.Second)
//	if err != nil {
//	    log.Printf("Function shutdown timeout: %v", err)
//	}
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

// Go spawns a new tracked goroutine with advanced lifecycle management.
// The goroutine is automatically tracked, monitored via metrics, and cleaned up on completion.
//
// Parameters:
//   - functionName: Logical name for the goroutine (used for grouping and metrics)
//   - workerFunc: The function to execute in the goroutine (receives context)
//   - opts: Optional configuration (timeout, panic recovery, wait group)
//
// Available Options:
//   - WithTimeout(duration): Auto-cancels goroutine after specified duration
//   - WithPanicRecovery(bool): Enable/disable panic recovery (enabled by default)
//   - AddToWaitGroup(name): Add to function-level wait group for coordinated shutdown
//
// Goroutine Lifecycle:
//  1. Creates child context derived from local manager's context
//  2. Applies timeout if specified
//  3. Creates routine instance with unique ID
//  4. Adds to tracking map and wait groups
//  5. Spawns goroutine with context
//  6. Executes worker function
//  7. On completion/panic: records metrics, decrements wait groups, closes done channel
//  8. Automatically removes from tracking map to prevent memory leaks
//
// Context Cancellation:
//
//	The context passed to workerFunc will be cancelled when:
//	- Timeout expires (if WithTimeout was used)
//	- Local manager shuts down
//	- App manager shuts down
//	- Global manager shuts down
//	- Manual cancellation via CancelRoutine()
//
// Returns:
//   - error: nil on success, error if local manager not found
//
// Example:
//
//	err := localMgr.Go("http-handler", func(ctx context.Context) error {
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return nil  // Graceful shutdown
//	        default:
//	            // Do work
//	        }
//	    }
//	}, local.WithTimeout(5*time.Minute), local.AddToWaitGroup("handlers"))
func (LM *LocalManagerStruct) Go(functionName string, workerFunc func(ctx context.Context) error, opts ...interfaces.GoroutineOption) error {
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

// spawnGoroutine is the internal implementation for spawning and tracking goroutines.
// It handles context creation, wait group management, panic recovery, and cleanup.
//
// Parameters:
//   - functionName: Logical name for grouping and metrics
//   - workerFunc: The function to execute in the goroutine
//   - opts: Configuration options (timeout, panic recovery, wait group)
//
// Implementation Details:
//   - Creates child context with optional timeout
//   - Increments function wait group (if specified) BEFORE spawning
//   - Increments local manager wait group for safe shutdown
//   - Spawns goroutine with panic recovery (if enabled)
//   - Records creation, completion, and error metrics
//   - Automatically cleans up on completion (wait groups, tracking map, done channel)
//   - Ensures no goroutine leaks via RemoveRoutine() call
//
// Returns:
//   - error: nil on successful spawn, error if local manager not found
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

// GetAllGoroutines retrieves all tracked goroutines in this local manager.
// This returns a slice of all routine instances managed by this local manager.
//
// Returns:
//   - []*types.Routine: Slice of all goroutines in this local manager
//   - error: Returns error if local manager is not found
//
// Note: Creates a new slice from the internal map. For just counting, use GetGoroutineCount() instead.
//
// Example:
//
//	routines, err := localMgr.GetAllGoroutines()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	}
//	for _, routine := range routines {
//	    fmt.Printf("Routine: %s, Function: %s\n", routine.GetID(), routine.GetFunctionName())
//	}
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

// GetGoroutineCount returns the number of tracked goroutines in this local manager.
// This uses an atomic counter for O(1) lock-free reads, making it highly efficient.
//
// Returns:
//   - int: Count of active goroutines
//     Returns 0 if local manager is not found
//
// Performance: O(1) with atomic read, no locking required
//
// Example:
//
//	count := localMgr.GetGoroutineCount()
//	log.Printf("Active goroutines: %d", count)
func (LM *LocalManagerStruct) GetGoroutineCount() int {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return 0
	}
	return localManager.GetRoutineCount()
}

// NewFunctionWaitGroup creates or retrieves a function-level wait group.
// Function wait groups allow coordinated shutdown of all goroutines with the same function name.
//
// Parameters:
//   - ctx: Context for the operation (currently unused, reserved for future use)
//   - functionName: Name of the function to create/get wait group for
//
// Behavior:
//   - If wait group already exists: returns existing wait group
//   - If not exists: creates new wait group and returns it
//   - Thread-safe: protected by local manager's mutex
//
// Returns:
//   - *sync.WaitGroup: The function wait group
//   - error: Returns error if local manager is not found
//
// Example:
//
//	wg, err := localMgr.NewFunctionWaitGroup(ctx, "worker")
//	if err != nil {
//	    log.Printf("Error creating wait group: %v", err)
//	}
//	// Use this WaitGroup is typically handled automatically by Go() with AddToWaitGroup option
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

// GetRoutinesByFunctionName retrieves all goroutines with a specific function name.
// This allows filtering goroutines by their logical grouping.
//
// Parameters:
//   - functionName: The function name to filter by
//
// Returns:
//   - []*types.Routine: Slice of all goroutines matching the function name
//   - error: Returns error if local manager is not found
//
// Example:
//
//	workers, err := localMgr.GetRoutinesByFunctionName("worker")
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	}
//	log.Printf("Found %d worker goroutines", len(workers))
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
