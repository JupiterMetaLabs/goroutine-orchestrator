package Interface

import (
	"context"
	"sync"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

// Initializer initializes the manager
type Initializer interface {
	Init() error
}

// Shutdowner handles shutdown of the manager
type Shutdowner interface {
	Shutdown(safe bool) error
}

// MetadataManager handles metadata of the Global manager
type MetadataManager interface {
	// NewMetadata() *types.Metadata
	GetMetadata() (*types.Metadata, error)
	UpdateMetadata(flag string, value interface{}) (*types.Metadata, error)
}

// GoroutineSpawner spawns and tracks goroutines
type GoroutineSpawner interface {
	Go(functionName string, workerFunc func(ctx context.Context) error) error
	GoWithWaitGroup(functionName string, workerFunc func(ctx context.Context) error) error
}

// FunctionShutdowner handles shutdown of specific functions
type FunctionShutdowner interface {
	ShutdownFunction(functionName string, timeout time.Duration) error
}

// GoroutineLister lists all tracked goroutines
type GoroutineLister interface {
	GetAllGoroutines() ([]*types.Routine, error)
	GetGoroutineCount() int
}

// LocalManagerLister lists all local managers
type LocalManagerLister interface {
	GetAllLocalManagers() ([]*types.LocalManager, error)
	GetLocalManagerCount() int
}

// AppManagerLister lists all app managers
type AppManagerLister interface {
	GetAllAppManagers() ([]*types.AppManager, error)
	GetAppManagerCount() int
}

// AppManagerCreator creates new app managers
type AppManagerCreator interface {
	CreateApp() (*types.AppManager, error)
}

// LocalManagerCreator creates new local managers
type LocalManagerCreator interface {
	CreateLocal(localName string) (*types.LocalManager, error)
}

type FunctionWaitGroupCreator interface {
	NewFunctionWaitGroup(ctx context.Context, functionName string) (*sync.WaitGroup, error)
}

// FunctionWaitGroupManager manages function-level wait groups
type FunctionWaitGroupManager interface {
	WaitForFunction(functionName string) error
	WaitForFunctionWithTimeout(functionName string, timeout time.Duration) bool
	GetFunctionGoroutineCount(functionName string) int
}

// RoutineManager defines methods for managing individual routines
type RoutineManager interface {
	CancelRoutine(routineID string) error
	WaitForRoutine(routineID string, timeout time.Duration) bool
	IsRoutineDone(routineID string) bool
	GetRoutineContext(routineID string) context.Context
	GetRoutineStartedAt(routineID string) int64
	GetRoutineUptime(routineID string) time.Duration
	IsRoutineContextCancelled(routineID string) bool
	GetRoutine(routineID string) (*types.Routine, error)
}

// ----------------------
// Composed interfaces
// ----------------------

// GlobalGoroutineManagerInterface defines the complete interface for global manager
type GlobalGoroutineManagerInterface interface {
	Initializer
	Shutdowner

	MetadataManager

	AppManagerLister

	LocalManagerLister

	GoroutineLister
}

// AppGoroutineManagerInterface defines the complete interface for app manager
type AppGoroutineManagerInterface interface {
	Shutdowner

	AppManagerCreator

	LocalManagerLister

	GoroutineLister
}

// LocalGoroutineManagerInterface defines the complete interface for local manager
type LocalGoroutineManagerInterface interface {
	Shutdowner
	FunctionShutdowner

	LocalManagerCreator

	GoroutineSpawner

	GoroutineLister

	FunctionWaitGroupCreator
	FunctionWaitGroupManager
}
