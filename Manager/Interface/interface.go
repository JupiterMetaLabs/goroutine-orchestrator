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

// GoroutineSpawner spawns and tracks goroutines
type GoroutineSpawner interface {
	Go(functionName string, workerFunc func(ctx context.Context) error) error
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

// ----------------------
// Composed interfaces
// ----------------------

// GlobalGoroutineManagerInterface defines the complete interface for global manager
type GlobalGoroutineManagerInterface interface {
	Initializer
	Shutdowner

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
	LocalManagerCreator

	FunctionShutdowner
	Shutdowner

	GoroutineSpawner
	GoroutineLister

	FunctionWaitGroupCreator
}
