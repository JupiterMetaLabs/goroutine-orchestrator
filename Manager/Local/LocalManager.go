package Local

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
	"github.com/neerajchowdary889/GoRoutinesManager/types/Errors"
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
	// First get the app manager
	appManager, err := types.GetAppManager(LM.AppName)
	if err != nil {
		return nil, err
	}

	// Directly call the CreateLocal method of the app manager
	// CreateLocal function will handle the checking and creation of the local manager
	localManager, err := appManager.CreateLocal(localName)
	switch err {
	case Errors.ErrLocalManagerNotFound:
		return nil, fmt.Errorf("%w: %s", Errors.ErrLocalManagerNotFound, localName)
	case Errors.WrngLocalManagerAlreadyExists:
		// Return the existing local manager and also return error as nil
		return localManager, nil
	default:
		// Fill the structs
		localManager.SetLocalContext().
			SetLocalMutex().
			SetLocalWaitGroup()
	}
	return localManager, nil
}

// Shutdowner
func (LM *LocalManagerStruct) Shutdown(safe bool) error {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return err
	}

	if safe {
		// Safe shutdown: try graceful shutdown first, then force cancel hanging goroutines

		// Step 1: Get all unique function names
		routines, err := LM.GetAllGoroutines()
		if err != nil {
			return err
		}

		functionNames := make(map[string]bool)
		for _, routine := range routines {
			functionNames[routine.GetFunctionName()] = true
		}

		// Step 2: Try to shutdown each function gracefully with timeout
		shutdownTimeout := types.ShutdownTimeout
		for functionName := range functionNames {
			// Try graceful shutdown with timeout
			LM.ShutdownFunction(functionName, shutdownTimeout)
			// Note: ShutdownFunction already handles cancellation and waiting
		}

		// Step 3: Wait for main wait group with timeout
		done := make(chan struct{})
		go func() {
			if localManager.Wg != nil {
				localManager.Wg.Wait()
			}
			close(done)
		}()

		// Wait with timeout
		select {
		case <-done:
			// All goroutines completed gracefully
			return nil
		case <-time.After(shutdownTimeout):
			// Timeout - some goroutines are still hanging
			// Fall through to force cancel
		}

		// Step 4: Force cancel any remaining hanging goroutines
		remainingRoutines, err := LM.GetAllGoroutines()
		if err == nil {
			for _, routine := range remainingRoutines {
				cancel := routine.GetCancel()
				if cancel != nil {
					cancel()
				}
			}
		}

		// Cancel the local manager's context
		if localManager.Cancel != nil {
			localManager.Cancel()
		}

	} else {
		// Unsafe shutdown: cancel all contexts immediately
		// Get all routines and cancel their contexts
		routines, err := LM.GetAllGoroutines()
		if err != nil {
			return err
		}

		// Cancel all routine contexts
		for _, routine := range routines {
			cancel := routine.GetCancel()
			if cancel != nil {
				cancel()
			}
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
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return err
	}

	// Get all routines
	routines, err := LM.GetAllGoroutines()
	if err != nil {
		return err
	}

	// Cancel all routines with this function name
	for _, routine := range routines {
		if routine.GetFunctionName() == functionName {
			cancel := routine.GetCancel()
			if cancel != nil {
				cancel()
			}
		}
	}

	// Wait for completion with timeout
	completed := LM.WaitForFunctionWithTimeout(functionName, timeout)
	if !completed {
		return fmt.Errorf("shutdown timeout for function: %s", functionName)
	}

	// Clean up the wait group
	localManager.RemoveFunctionWg(functionName)

	return nil
}

// GoroutineSpawner
// Go spawns a new goroutine, tracks it in the LocalManager, and returns the routine ID.
// The goroutine is spawned with a context derived from the LocalManager's parent context.
// The done channel is closed when the goroutine completes.
func (LM *LocalManagerStruct) Go(functionName string, workerFunc func(ctx context.Context) error) error {
	return LM.spawnGoroutine(functionName, workerFunc, false)
}

// GoWithWaitGroup spawns a goroutine and automatically manages the function wait group.
// This is a convenience method that calls wg.Add(1) before spawning and wg.Done() after completion.
// Use this when you want automatic wait group management.
// For manual control, use Go() and manage the wait group yourself via NewFunctionWaitGroup().
func (LM *LocalManagerStruct) GoWithWaitGroup(functionName string, workerFunc func(ctx context.Context) error) error {
	return LM.spawnGoroutine(functionName, workerFunc, true)
}

// spawnGoroutine is the internal implementation for spawning goroutines.
// If useWaitGroup is true, it automatically manages Add/Done for the function wait group.
func (LM *LocalManagerStruct) spawnGoroutine(functionName string, workerFunc func(ctx context.Context) error, useWaitGroup bool) error {
	// Get the types.LocalManager instance
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return err
	}

	var wg *sync.WaitGroup
	if useWaitGroup {
		// Get or create function wait group
		wg, err = LM.NewFunctionWaitGroup(context.Background(), functionName)
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

	// Create the done channel (bidirectional, buffered size 1)
	// This allows non-blocking close even if nothing is reading
	doneChan := make(chan struct{}, 1)

	// Create a new Routine instance
	routine := types.NewGoRoutine(functionName).
		SetContext(routineCtx).
		SetCancel(cancel).
		SetDone(doneChan) // Override the channel created in NewGoRoutine

	// Add routine to LocalManager for tracking
	localManager.AddRoutine(routine)

	// Spawn the goroutine
	go func() {
		defer func() {
			if useWaitGroup && wg != nil {
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
		}()

		// Execute the worker function with the routine's context
		_ = workerFunc(routineCtx)
	}()

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

	// Retrieve and return the newly created wait group
	return localManager.GetFunctionWg(functionName)
}
