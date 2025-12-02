package types

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/JupiterMetaLabs/goroutine-orchestrator/ctxo"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/errors"
)

const (
	Prefix_LocalManager = "LocalManager."
)

func newLocalManager(localName string, appName string) *LocalManager {
	if IsIntilized().Local(localName, appName) {
		LocalManager, err := NewAppManager(appName).GetLocalManager(localName)
		if err != nil {
			return nil
		}
		return LocalManager
	}

	LocalManager := &LocalManager{
		LocalName:   localName,
		Routines:    make(map[string]*Routine),
		FunctionWgs: make(map[string]*sync.WaitGroup), // Initialize FunctionWgs map
		Wg:          &sync.WaitGroup{},                // Initialize wait group for safe shutdown
	}

	// Add the local manager to the app manager
	SetLocalManager(appName, localName, LocalManager)

	return LocalManager
}

// Lock APIs
// LockAppReadMutex locks the app read mutex for the app manager - This is used to read the app manager's data
func (LM *LocalManager) lockLocalReadMutex() {
	if LM.localMu == nil {
		LM.SetLocalMutex()
	}
	LM.localMu.RLock()
}

// UnlockAppReadMutex unlocks the app read mutex for the app manager - This is used to read the app manager's data
func (LM *LocalManager) unlockLocalReadMutex() {
	if LM.localMu == nil {
		LM.SetLocalMutex()
	}
	LM.localMu.RUnlock()
}

// LockAppWriteMutex locks the app write mutex for the app manager - This is used to write the app manager's data
func (LM *LocalManager) lockLocalWriteMutex() {
	if LM.localMu == nil {
		LM.SetLocalMutex()
	}
	LM.localMu.Lock()
}

// UnlockAppWriteMutex unlocks the app write mutex for the app manager - This is used to write the app manager's data
func (LM *LocalManager) unlockLocalWriteMutex() {
	if LM.localMu == nil {
		LM.SetLocalMutex()
	}
	LM.localMu.Unlock()
}

// Set the Local mutex to the local manager
func (LM *LocalManager) SetLocalMutex() *LocalManager {
	LM.localMu = &sync.RWMutex{}
	return LM
}

// >>> Set APIs
// SetLocalName sets the name of the local manager
func (LM *LocalManager) SetLocalName(localName string) *LocalManager {
	// Lock and update
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()
	LM.LocalName = localName
	return LM
}

// SetLocalContext sets the context for the local manager
func (LM *LocalManager) SetLocalContext() *LocalManager {
	// Lock and update
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()
	ctx := ctxo.GetAppContext(Prefix_LocalManager + LM.LocalName).Get()
	Done := func() {
		ctxo.GetAppContext(Prefix_LocalManager + LM.LocalName).Done(ctx)
	}
	LM.Ctx = ctx
	LM.Cancel = Done
	return LM
}

// SetLocalWaitGroup sets the wait group for the local manager
func (LM *LocalManager) SetLocalWaitGroup() *LocalManager {
	// Lock and update
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()
	LM.Wg = &sync.WaitGroup{}
	return LM
}

// SetParentContext sets the parent context for the local manager
func (LM *LocalManager) SetParentContext(ctx context.Context) *LocalManager {
	// Lock and update
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()

	LM.ParentCtx = ctx
	return LM
}

// SpawnChild sets the child context for the local manager
func (LM *LocalManager) SpawnChild() (context.Context, context.CancelFunc) {
	// Lock and update
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()

	ctx, cancel := ctxo.SpawnChild(LM.Ctx)
	return ctx, cancel
}

// AddRoutine adds a new routine to the local manager
func (LM *LocalManager) AddRoutine(routine *Routine) *LocalManager {
	// Lock and update
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()

	LM.Routines[routine.ID] = routine
	// Atomically increment routine count for lock-free reads
	atomic.AddInt64(&LM.routineCount, 1)
	return LM
}

// RemoveRoutine removes a routine from the local manager
func (LM *LocalManager) RemoveRoutine(routine *Routine, safe bool) *LocalManager {

	// Lock -> remove the routine -> unlock
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()

	// Cancel the routine's context to signal it to stop
	if routine.Cancel != nil {
		routine.Cancel()
	}

	// TODO: safe or unsafe terminate is based on the flag

	// Remove from the map
	if _, exists := LM.Routines[routine.ID]; exists {
		delete(LM.Routines, routine.ID)
		// Atomically decrement routine count for lock-free reads
		atomic.AddInt64(&LM.routineCount, -1)
	}
	return LM
}

// AddFunctionWg adds a new function wait group to the local manager
func (LM *LocalManager) AddFunctionWg(functionName string) *LocalManager {
	// Lock -> add the function wg -> unlock
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()

	LM.FunctionWgs[functionName] = &sync.WaitGroup{}
	return LM
}

// RemoveFunctionWg removes a function wait group from the local manager
func (LM *LocalManager) RemoveFunctionWg(functionName string) *LocalManager {

	// Lock -> remove the function wg -> unlock
	LM.lockLocalWriteMutex()
	defer LM.unlockLocalWriteMutex()

	delete(LM.FunctionWgs, functionName)
	return LM
}

// >>> Get APIs
// GetRoutine gets a specific routine for the local manager
func (LM *LocalManager) GetRoutine(routineID string) (*Routine, error) {
	LM.lockLocalReadMutex()
	defer LM.unlockLocalReadMutex()

	if _, ok := LM.Routines[routineID]; !ok {
		return nil, fmt.Errorf("%w: %s", errors.ErrRoutineNotFound, routineID)
	}
	return LM.Routines[routineID], nil
}

// GetRoutines gets all the routines for the local manager
func (LM *LocalManager) GetRoutines() map[string]*Routine {
	LM.lockLocalReadMutex()
	defer LM.unlockLocalReadMutex()

	routinesCopy := make(map[string]*Routine, len(LM.Routines))
	for k, v := range LM.Routines {
		routinesCopy[k] = v
	}
	return routinesCopy
}

// GetLocalContext gets the context for the local manager
func (LM *LocalManager) GetLocalContext() (context.Context, context.CancelFunc) {
	return LM.Ctx, LM.Cancel
}

// GetFunctionWg gets a specific function wait group for the local manager
func (LM *LocalManager) GetFunctionWg(functionName string) (*sync.WaitGroup, error) {
	LM.lockLocalReadMutex()
	defer LM.unlockLocalReadMutex()

	if _, ok := LM.FunctionWgs[functionName]; !ok {
		return nil, fmt.Errorf("%w: %s", errors.ErrFunctionWgNotFound, functionName)
	}
	return LM.FunctionWgs[functionName], nil
}

// GetRoutineCount gets the number of routines for the local manager
// Uses atomic read for lock-free performance on high-frequency calls
// Falls back to mutex-protected len() if atomic value is inconsistent (shouldn't happen)
func (LM *LocalManager) GetRoutineCount() int {
	// Lock-free read using atomic counter
	count := int(atomic.LoadInt64(&LM.routineCount))

	// Sanity check: if count is negative, something went wrong - use mutex path
	// This should never happen in normal operation, but provides safety
	if count < 0 {
		LM.lockLocalReadMutex()
		defer LM.unlockLocalReadMutex()
		// Reset atomic counter to actual value
		actualCount := len(LM.Routines)
		atomic.StoreInt64(&LM.routineCount, int64(actualCount))
		return actualCount
	}

	return count
}

// GetFunctionWgCount gets the number of function wait groups for the local manager
func (LM *LocalManager) GetFunctionWgCount() int {
	LM.lockLocalReadMutex()
	defer LM.unlockLocalReadMutex()
	return len(LM.FunctionWgs)
}

// GetLocalName gets the name of the local manager
func (LM *LocalManager) GetLocalName() string {
	return LM.LocalName
}

// GetParentContext gets the parent context for the local manager
func (LM *LocalManager) GetParentContext() context.Context {
	return LM.ParentCtx
}

// GetLocalWaitGroup gets the wait group for the local manager
func (LM *LocalManager) GetLocalWaitGroup() *sync.WaitGroup {
	LM.lockLocalReadMutex()
	defer LM.unlockLocalReadMutex()
	return LM.Wg
}
