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
	switch err{
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
	//TODO
	return nil
}

// FunctionShutdowner
func (LM *LocalManagerStruct) ShutdownFunction(functionName string, timeout time.Duration) error {
	//TODO
	return nil
}

// GoroutineSpawner
// Go spawns a new goroutine, tracks it in the LocalManager, and returns the routine ID.
// The goroutine is spawned with a context derived from the LocalManager's parent context.
// The done channel is closed when the goroutine completes.
func (LM *LocalManagerStruct) Go(functionName string, workerFunc func(ctx context.Context) error) error {
	// Get the types.LocalManager instance
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return err
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
	//TODO
	return nil, nil
}

// Routine management methods - these operate on individual routines by ID

// CancelRoutine cancels a routine's context by its ID.
// Returns an error if the routine is not found.
func (LM *LocalManagerStruct) CancelRoutine(routineID string) error {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return err
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		return err
	}

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
func (LM *LocalManagerStruct) GetRoutineContext(routineID string) context.Context {
	localManager, err := types.GetLocalManager(LM.AppName, LM.LocalName)
	if err != nil {
		return context.Background()
	}

	routine, err := localManager.GetRoutine(routineID)
	if err != nil {
		return context.Background()
	}

	ctx := routine.GetContext()
	if ctx == nil {
		return context.Background()
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
