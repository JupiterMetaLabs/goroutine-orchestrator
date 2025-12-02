package Local

import (
	"context"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Context"
	"github.com/neerajchowdary889/GoRoutinesManager/metrics"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

// Routine management methods - these operate on individual routines by ID

// CancelRoutine cancels a routine's context by its ID.
// Returns an error if the routine is not found.
func (LM *LocalManagerStruct) CancelRoutine(routineID string) error {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		metrics.RecordOperationError("goroutine", "cancel", "get_local_manager_failed")
		return err
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		metrics.RecordOperationError("goroutine", "cancel", "routine_not_found")
		return err
	}

	functionName := routine.GetFunctionName()
	// Record operation
	metrics.RecordGoroutineOperation("cancel", LM.AppName, LM.LocalName, functionName)

	cancel := routine.GetCancel()
	if cancel != nil {
		cancel()
	}
	return nil
}

// WaitForRoutine blocks until the routine's done channel is signaled or the timeout expires.
// Returns true if the routine completed, false if timeout occurred or routine not found.
func (LM *LocalManagerStruct) WaitForRoutine(routineID string, timeout time.Duration) bool {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return false
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		return false
	}

	doneChan := routine.DoneChan()
	if doneChan == nil {
		return false
	}

	select {
	case <-doneChan:
		return true
	case <-time.After(timeout):
		return false
	}
}

// IsRoutineDone checks if a routine's done channel has been signaled.
// Returns false if routine is not found or done channel is nil.
func (LM *LocalManagerStruct) IsRoutineDone(routineID string) bool {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return false
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		return false
	}

	doneChan := routine.DoneChan()
	if doneChan == nil {
		return false
	}

	select {
	case <-doneChan:
		return true
	default:
		return false
	}
}

// GetRoutineContext returns the context associated with a routine by ID.
// Returns context.Background() if routine is not found or context is nil.
// Instead of returning context.Background(), spawn a background using Context.Spawnchild(localparent ctx)
func (LM *LocalManagerStruct) GetRoutineContext(routineID string) context.Context {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		ctx, _ := Context.SpawnChild(localManager.ParentCtx)
		return ctx
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		ctx, _ := Context.SpawnChild(localManager.ParentCtx)
		return ctx
	}

	ctx := routine.GetContext()
	if ctx == nil {
		ctx, _ := Context.SpawnChild(localManager.ParentCtx)
		return ctx
	}
	return ctx
}

// GetRoutineStartedAt returns the timestamp when a routine was started.
// Returns 0 if routine is not found.
func (LM *LocalManagerStruct) GetRoutineStartedAt(routineID string) int64 {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return 0
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		return 0
	}

	return routine.GetStartedAt()
}

// GetRoutineUptime returns the duration a routine has been running.
// Returns 0 if routine is not found or not started.
func (LM *LocalManagerStruct) GetRoutineUptime(routineID string) time.Duration {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return 0
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		return 0
	}

	startedAt := routine.GetStartedAt()
	if startedAt == 0 {
		return 0
	}

	now := time.Now().UnixNano()
	return time.Duration(now - startedAt)
}

// IsRoutineContextCancelled checks if a routine's context has been cancelled.
// Returns false if routine is not found or context is nil.
func (LM *LocalManagerStruct) IsRoutineContextCancelled(routineID string) bool {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return false
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		return false
	}

	ctx := routine.GetContext()
	if ctx == nil {
		return false
	}

	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// GetRoutine returns a routine by its ID.
// Returns an error if the routine is not found.
func (LM *LocalManagerStruct) GetRoutine(routineID string) (*types.Routine, error) {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return nil, err
	}
	return localManager.GetRoutine(routineID)
}

// WaitForFunction waits for all goroutines of a specific function to complete.
func (LM *LocalManagerStruct) WaitForFunction(functionName string) error {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return err
	}

	wg, err := localManager.GetFunctionWg(functionName)
	if err != nil {
		return err // No wait group for this function
	}

	wg.Wait()
	return nil
}

// WaitForFunctionWithTimeout waits for all goroutines of a function with a timeout.
// Returns true if all completed, false if timeout occurred.
//
// Note: If timeout occurs, the internal goroutine will continue waiting until
// the WaitGroup completes. This is acceptable because:
// 1. The goroutine will eventually exit when WaitGroup is done
// 2. No resources are held except memory
// 3. The function correctly returns false on timeout
func (LM *LocalManagerStruct) WaitForFunctionWithTimeout(functionName string, timeout time.Duration) bool {
	done := make(chan struct{}, 1) // Buffered to prevent goroutine leak if timeout occurs

	go func() {
		// WaitForFunction blocks on wg.Wait() which cannot be cancelled.
		// If timeout occurs, this goroutine will continue running until
		// the WaitGroup completes (when all goroutines finish).
		// This is acceptable - the goroutine will eventually exit.
		err := LM.WaitForFunction(functionName)
		if err == nil {
			// WaitGroup completed - signal success (non-blocking due to buffered channel)
			select {
			case done <- struct{}{}:
			default:
				// Timeout already occurred, but WaitGroup completed - no-op
			}
		}
		// Note: We don't close the channel here because the select might
		// have already returned on timeout. Closing would cause a panic
		// if someone tries to read from it again.
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		// Timeout occurred - return false immediately
		// The goroutine will continue running until WaitGroup completes,
		// but this is acceptable as it will eventually exit.
		return false
	}
}

// GetFunctionGoroutineCount returns the number of goroutines for a specific function.
func (LM *LocalManagerStruct) GetFunctionGoroutineCount(functionName string) int {
	routines, err := LM.GetAllGoroutines()
	if err != nil {
		return 0
	}

	count := 0
	for _, routine := range routines {
		if routine.GetFunctionName() == functionName {
			count++
		}
	}
	return count
}
