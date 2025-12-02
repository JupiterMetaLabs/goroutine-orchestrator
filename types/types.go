package types

import (
	"context"
	"sync"
	"time"
)

// Default Values
var (
	// Default timeout is 10 seconds - can be changed using Metadata
	ShutdownTimeout = 10 * time.Second
	// Default update interval is 5 seconds - can be changed using Metadata
	UpdateInterval = 5 * time.Second
)

// Singleton pattern to not repeat the same managers again
var (
	Global *GlobalManager
)

// GlobalManager manages all app-level managers
type GlobalManager struct {
	globalMu    *sync.RWMutex
	AppManagers map[string]*AppManager
	Ctx         context.Context
	Cancel      context.CancelFunc
	Wg          *sync.WaitGroup
	Metadata    *Metadata
}

// AppManager manages local-level managers for a specific app/module
type AppManager struct {
	appMu         *sync.RWMutex
	AppName       string
	LocalManagers map[string]*LocalManager
	Ctx           context.Context
	Cancel        context.CancelFunc
	Wg            *sync.WaitGroup
	ParentCtx     context.Context
}

// LocalManager manages goroutines for a specific file/module within an app
type LocalManager struct {
	localMu     *sync.RWMutex
	LocalName   string
	Routines    map[string]*Routine
	Ctx         context.Context
	Cancel      context.CancelFunc
	Wg          *sync.WaitGroup
	FunctionWgs map[string]*sync.WaitGroup // Per function name for selective shutdown
	ParentCtx   context.Context
	// Atomic counter for lock-free reads of routine count
	// Updated atomically when routines are added/removed
	routineCount int64 // Use sync/atomic for operations
}

// Routine represents a tracked goroutine
type Routine struct {
	ID           string
	FunctionName string
	Ctx          context.Context
	Cancel       context.CancelFunc
	Done         <-chan struct{}
	StartedAt    int64 // Unix timestamp or monotonic time
}

type Metadata struct {
	metadataMu *sync.RWMutex
	MaxRoutines     int
	Metrics         bool
	MetricsURL      string
	UpdateInterval  time.Duration
	ShutdownTimeout time.Duration
}
