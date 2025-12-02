# GoRoutinesManager Usage Guide

This document provides comprehensive guidance on using the GoRoutinesManager package. It covers initialization, common patterns, and best practices for each manager type.

---

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
- [GlobalManager](#globalmanager)
- [AppManager](#appmanager)
- [LocalManager](#localmanager)
- [Common Patterns](#common-patterns)
- [Shutdown Strategies](#shutdown-strategies)
- [Error Handling](#error-handling)
- [Advanced Usage](#advanced-usage)

---

## Overview

GoRoutinesManager provides three levels of managers:

1. **GlobalManager** - Singleton that manages the entire application lifecycle
2. **AppManager** - Manages a logical application or service
3. **LocalManager** - Manages goroutines for a specific module or file

The typical workflow is:
1. Initialize GlobalManager
2. Create one or more AppManagers
3. Create LocalManagers within each AppManager
4. Spawn goroutines using LocalManager

---

## Getting Started

### Import the Package

```go
import (
    "github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
    "github.com/neerajchowdary889/GoRoutinesManager/Manager/App"
    "github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
)
```

### Basic Initialization Flow

The recommended initialization order:

1. Initialize GlobalManager (sets up signal handling)
2. Create AppManager(s) for your applications
3. Create LocalManager(s) for your modules
4. Spawn goroutines using LocalManager

---

## GlobalManager

The GlobalManager is a singleton that orchestrates the entire application lifecycle.

### Initialization

**Function:** `Init() error`

Initializes the global manager and sets up signal handling for SIGINT and SIGTERM. This should be called early in your application startup.

**Example:**
```go
globalMgr := Global.NewGlobalManager()
if err := globalMgr.Init(); err != nil {
    log.Fatalf("Failed to initialize global manager: %v", err)
}
```

**Key Points:**
- Idempotent: Safe to call multiple times
- Automatically sets up signal handlers
- Creates global context for the process
- Initializes metadata with default values

### Configuration (Metadata)

**Function:** `UpdateMetadata(flag string, value interface{}) (*types.Metadata, error)`

Configure global settings such as shutdown timeout, metrics, and limits.

**Available Flags:**

- `SET_SHUTDOWN_TIMEOUT` - Set graceful shutdown timeout (time.Duration)
- `SET_METRICS_URL` - Enable metrics and set URL (string or []interface{})
- `SET_MAX_ROUTINES` - Set maximum routines limit (int)
- `SET_UPDATE_INTERVAL` - Set metrics update interval (time.Duration)

**Examples:**

```go
// Set shutdown timeout to 30 seconds
globalMgr.UpdateMetadata("SET_SHUTDOWN_TIMEOUT", 30*time.Second)

// Enable metrics with URL
globalMgr.UpdateMetadata("SET_METRICS_URL", ":9090")

// Enable metrics with custom interval
globalMgr.UpdateMetadata("SET_METRICS_URL", []interface{}{true, ":9090", 10*time.Second})

// Set maximum routines limit
globalMgr.UpdateMetadata("SET_MAX_ROUTINES", 1000)

// Set metrics update interval
globalMgr.UpdateMetadata("SET_UPDATE_INTERVAL", 5*time.Second)
```

### Querying State

**Functions:**
- `GetAllAppManagers() ([]*types.AppManager, error)` - Get all app managers
- `GetAppManagerCount() int` - Get count of app managers
- `GetAllLocalManagers() ([]*types.LocalManager, error)` - Get all local managers
- `GetLocalManagerCount() int` - Get total count of local managers
- `GetAllGoroutines() ([]*types.Routine, error)` - Get all tracked goroutines
- `GetGoroutineCount() int` - Get total count of tracked goroutines
- `GetMetadata() (*types.Metadata, error)` - Get current metadata

**Example:**
```go
// Get all app managers
apps, err := globalMgr.GetAllAppManagers()
if err != nil {
    log.Printf("Error getting app managers: %v", err)
}

// Get total goroutine count
count := globalMgr.GetGoroutineCount()
log.Printf("Total goroutines: %d", count)
```

### Shutdown

**Function:** `Shutdown(safe bool) error`

Shuts down all app managers and the global context.

- `safe=true`: Graceful shutdown - waits for goroutines to complete, then force cancels if timeout
- `safe=false`: Immediate shutdown - cancels all contexts immediately

**Example:**
```go
// Graceful shutdown
if err := globalMgr.Shutdown(true); err != nil {
    log.Printf("Shutdown error: %v", err)
}

// Immediate shutdown
globalMgr.Shutdown(false)
```

**Note:** The global context automatically handles SIGINT/SIGTERM signals and triggers shutdown. You typically don't need to call `Shutdown()` manually unless you want to shutdown programmatically.

---

## AppManager

AppManager manages a logical application or service within your system.

### Creation

**Function:** `CreateApp() (*types.AppManager, error)`

Creates and registers an app manager. If the global manager is not initialized, it will be auto-initialized.

**Example:**
```go
appMgr := App.NewAppManager("api-server")
app, err := appMgr.CreateApp()
if err != nil {
    log.Fatalf("Failed to create app: %v", err)
}
```

**Key Points:**
- Idempotent: Returns existing app if already created
- Auto-initializes global manager if needed
- Creates app-level context derived from global context
- App name must be unique

### Creating Local Managers

**Function:** `CreateLocal(localName string) (*types.LocalManager, error)`

Creates a local manager within the app. This is typically done through the LocalManager interface, but can also be done directly.

**Example:**
```go
localMgr := Local.NewLocalManager("api-server", "handlers")
local, err := localMgr.CreateLocal("handlers")
if err != nil {
    log.Fatalf("Failed to create local manager: %v", err)
}
```

### Querying State

**Functions:**
- `GetAllLocalManagers() ([]*types.LocalManager, error)` - Get all local managers in the app
- `GetLocalManagerCount() int` - Get count of local managers
- `GetLocalManagerByName(localName string) (*types.LocalManager, error)` - Get specific local manager
- `GetAllGoroutines() ([]*types.Routine, error)` - Get all goroutines in the app
- `GetGoroutineCount() int` - Get count of goroutines in the app

**Example:**
```go
// Get all local managers
locals, err := appMgr.GetAllLocalManagers()
if err != nil {
    log.Printf("Error: %v", err)
}

// Get goroutine count for this app
count := appMgr.GetGoroutineCount()
log.Printf("App goroutines: %d", count)
```

### Shutdown

**Function:** `Shutdown(safe bool) error`

Shuts down all local managers within the app.

- `safe=true`: Graceful shutdown with timeout
- `safe=false`: Immediate cancellation

**Example:**
```go
// Shutdown this app gracefully
if err := appMgr.Shutdown(true); err != nil {
    log.Printf("App shutdown error: %v", err)
}
```

---

## LocalManager

LocalManager manages goroutines for a specific module or file within an app.

### Creation

**Function:** `CreateLocal(localName string) (*types.LocalManager, error)`

Creates a local manager. The app manager must exist first.

**Example:**
```go
// First create the app
appMgr := App.NewAppManager("api-server")
appMgr.CreateApp()

// Then create the local manager
localMgr := Local.NewLocalManager("api-server", "handlers")
local, err := localMgr.CreateLocal("handlers")
if err != nil {
    log.Fatalf("Failed to create local manager: %v", err)
}
```

**Key Points:**
- Idempotent: Returns existing local manager if already created
- Creates local-level context derived from app context
- Local name must be unique within the app

### Spawning Goroutines

**Function:** `Go(functionName string, workerFunc func(ctx context.Context) error, opts ...GoroutineOption) error`

Spawns a tracked goroutine. The worker function receives a context that will be cancelled when the local manager shuts down.

**Basic Example:**
```go
localMgr.Go("worker", func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil // Exit gracefully
        default:
            // Do work
            time.Sleep(1 * time.Second)
        }
    }
})
```

### Goroutine Options

You can configure goroutines using options:

#### WithTimeout

Sets a timeout for the goroutine. The context will be cancelled when the timeout expires.

```go
localMgr.Go("timeout-worker", func(ctx context.Context) error {
    // This will be cancelled after 5 seconds
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // Do work
        }
    }
}, Local.WithTimeout(5 * time.Second))
```

#### WithPanicRecovery

Enables or disables panic recovery. Enabled by default.

```go
// Panic recovery enabled (default)
localMgr.Go("safe-worker", func(ctx context.Context) error {
    panic("something went wrong") // Won't crash the app
})

// Disable panic recovery (not recommended)
localMgr.Go("unsafe-worker", func(ctx context.Context) error {
    // Panics will crash the goroutine
}, Local.WithPanicRecovery(false))
```

#### AddToWaitGroup

Adds the goroutine to a function-level wait group for coordinated shutdown.

```go
// Spawn multiple goroutines for the same function
for i := 0; i < 10; i++ {
    localMgr.Go("worker", func(ctx context.Context) error {
        // Do work
        return nil
    }, Local.AddToWaitGroup("worker"))
}

// Wait for all "worker" goroutines to complete
localMgr.WaitForFunction("worker")
```

### Function Wait Groups

Function wait groups allow you to coordinate multiple goroutines with the same function name.

**Functions:**
- `NewFunctionWaitGroup(ctx context.Context, functionName string) (*sync.WaitGroup, error)` - Create or get wait group
- `WaitForFunction(functionName string) error` - Wait for all goroutines of a function
- `WaitForFunctionWithTimeout(functionName string, timeout time.Duration) bool` - Wait with timeout
- `GetFunctionGoroutineCount(functionName string) int` - Get count of goroutines for a function

**Example:**
```go
// Spawn multiple workers
for i := 0; i < 5; i++ {
    localMgr.Go("worker", workerFunc, Local.AddToWaitGroup("worker"))
}

// Wait for all workers with timeout
completed := localMgr.WaitForFunctionWithTimeout("worker", 30*time.Second)
if !completed {
    log.Println("Workers did not complete within timeout")
}
```

### Selective Shutdown

**Function:** `ShutdownFunction(functionName string, timeout time.Duration) error`

Shuts down all goroutines of a specific function with a timeout.

**Example:**
```go
// Shutdown all "worker" goroutines
err := localMgr.ShutdownFunction("worker", 10*time.Second)
if err != nil {
    log.Printf("Shutdown timeout: %v", err)
}
```

### Routine Management

**Functions:**
- `GetAllGoroutines() ([]*types.Routine, error)` - Get all tracked goroutines
- `GetGoroutineCount() int` - Get count of tracked goroutines
- `GetRoutine(routineID string) (*types.Routine, error)` - Get specific routine
- `GetRoutinesByFunctionName(functionName string) ([]*types.Routine, error)` - Get routines by function name
- `CancelRoutine(routineID string) error` - Cancel a specific routine
- `WaitForRoutine(routineID string, timeout time.Duration) bool` - Wait for routine completion
- `IsRoutineDone(routineID string) bool` - Check if routine is done
- `GetRoutineContext(routineID string) context.Context` - Get routine's context
- `GetRoutineStartedAt(routineID string) int64` - Get start timestamp
- `GetRoutineUptime(routineID string) time.Duration` - Get uptime duration
- `IsRoutineContextCancelled(routineID string) bool` - Check if context is cancelled

**Example:**
```go
// Get all routines
routines, err := localMgr.GetAllGoroutines()
if err != nil {
    log.Printf("Error: %v", err)
}

// Get routines for a specific function
workerRoutines, err := localMgr.GetRoutinesByFunctionName("worker")
if err != nil {
    log.Printf("Error: %v", err)
}

// Cancel a specific routine
if err := localMgr.CancelRoutine("routine-id-123"); err != nil {
    log.Printf("Error cancelling routine: %v", err)
}

// Check routine status
if localMgr.IsRoutineDone("routine-id-123") {
    log.Println("Routine is done")
}
```

### Shutdown

**Function:** `Shutdown(safe bool) error`

Shuts down all goroutines in the local manager.

- `safe=true`: Graceful shutdown - attempts graceful shutdown, then force cancels if timeout
- `safe=false`: Immediate cancellation

**Example:**
```go
// Graceful shutdown
if err := localMgr.Shutdown(true); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

---

## Common Patterns

### Pattern 1: Simple Application Setup

```go
// Initialize global manager
globalMgr := Global.NewGlobalManager()
globalMgr.Init()

// Create app
appMgr := App.NewAppManager("my-app")
appMgr.CreateApp()

// Create local manager
localMgr := Local.NewLocalManager("my-app", "workers")
localMgr.CreateLocal("workers")

// Spawn goroutines
localMgr.Go("worker", func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // Do work
        }
    }
})

// Wait for shutdown signal (handled automatically by global context)
select {}
```

### Pattern 2: Multiple Apps

```go
globalMgr := Global.NewGlobalManager()
globalMgr.Init()

// Create multiple apps
apiApp := App.NewAppManager("api-server")
apiApp.CreateApp()

workerApp := App.NewAppManager("worker-pool")
workerApp.CreateApp()

// Setup each app independently
apiLocal := Local.NewLocalManager("api-server", "handlers")
apiLocal.CreateLocal("handlers")

workerLocal := Local.NewLocalManager("worker-pool", "jobs")
workerLocal.CreateLocal("jobs")
```

### Pattern 3: Function-Level Coordination

```go
localMgr := Local.NewLocalManager("my-app", "workers")
localMgr.CreateLocal("workers")

// Spawn multiple workers
for i := 0; i < 10; i++ {
    localMgr.Go("worker", func(ctx context.Context) error {
        // Do work
        return nil
    }, Local.AddToWaitGroup("worker"))
}

// Wait for all workers to complete
localMgr.WaitForFunction("worker")
```

### Pattern 4: Timeout Protection

```go
localMgr.Go("long-task", func(ctx context.Context) error {
    // This will be cancelled after 5 minutes
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // Long-running task
        }
    }
}, Local.WithTimeout(5 * time.Minute))
```

### Pattern 5: HTTP Server Integration

```go
globalMgr := Global.NewGlobalManager()
globalMgr.Init()

// Enable metrics
globalMgr.UpdateMetadata("SET_METRICS_URL", ":9090")

appMgr := App.NewAppManager("api-server")
appMgr.CreateApp()

localMgr := Local.NewLocalManager("api-server", "server")
localMgr.CreateLocal("server")

localMgr.Go("http-server", func(ctx context.Context) error {
    mux := http.NewServeMux()
    mux.Handle("/metrics", metrics.GetMetricsHandler())
    mux.HandleFunc("/api", apiHandler)
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
    
    // Shutdown server when context is cancelled
    go func() {
        <-ctx.Done()
        server.Shutdown(context.Background())
    }()
    
    return server.ListenAndServe()
})
```

---

## Shutdown Strategies

### Strategy 1: Automatic Signal Handling (Recommended)

The global context automatically handles SIGINT and SIGTERM signals. No manual shutdown code needed.

```go
globalMgr := Global.NewGlobalManager()
globalMgr.Init()

// ... setup your application ...

// Block forever - shutdown handled by signals
select {}
```

### Strategy 2: Programmatic Shutdown

Shutdown programmatically when needed.

```go
// Graceful shutdown
if err := globalMgr.Shutdown(true); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

### Strategy 3: Selective Shutdown

Shutdown specific components without affecting others.

```go
// Shutdown a specific function
localMgr.ShutdownFunction("worker", 10*time.Second)

// Shutdown a specific app
appMgr.Shutdown(true)

// Shutdown a specific local manager
localMgr.Shutdown(true)
```

### Strategy 4: Timeout Configuration

Configure shutdown timeouts based on your needs.

```go
// Set global shutdown timeout
globalMgr.UpdateMetadata("SET_SHUTDOWN_TIMEOUT", 30*time.Second)

// Safe shutdown will wait up to 30 seconds
globalMgr.Shutdown(true)
```

---

## Error Handling

### Initialization Errors

Always check for errors during initialization.

```go
globalMgr := Global.NewGlobalManager()
if err := globalMgr.Init(); err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}

appMgr := App.NewAppManager("my-app")
app, err := appMgr.CreateApp()
if err != nil {
    log.Fatalf("Failed to create app: %v", err)
}
```

### Shutdown Errors

Shutdown errors typically indicate timeouts or stuck goroutines.

```go
if err := localMgr.Shutdown(true); err != nil {
    log.Printf("Shutdown had issues: %v", err)
    // Some goroutines may have been force-cancelled
}
```

### Routine Errors

Worker functions should return errors for proper error handling.

```go
localMgr.Go("worker", func(ctx context.Context) error {
    if err := doWork(); err != nil {
        return fmt.Errorf("work failed: %w", err)
    }
    return nil
})
```

---

## Advanced Usage

### Custom Context Propagation

You can create custom child contexts for fine-grained control.

```go
// Get app context
appCtx, _ := appMgr.GetAppContext()

// Create custom child context with timeout
customCtx, cancel := context.WithTimeout(appCtx, 5*time.Second)
defer cancel()

// Use custom context in your code
```

### Metadata Management

Query and update metadata dynamically.

```go
// Get current metadata
metadata, err := globalMgr.GetMetadata()
if err != nil {
    log.Printf("Error: %v", err)
}

// Check if metrics are enabled
if metadata.GetMetrics() {
    log.Println("Metrics are enabled")
}

// Update shutdown timeout
globalMgr.UpdateMetadata("SET_SHUTDOWN_TIMEOUT", 60*time.Second)
```

### Routine Inspection

Inspect individual routines for debugging and monitoring.

```go
// Get all routines
routines, err := localMgr.GetAllGoroutines()
if err != nil {
    log.Printf("Error: %v", err)
}

for _, routine := range routines {
    log.Printf("Routine ID: %s", routine.GetID())
    log.Printf("Function: %s", routine.GetFunctionName())
    log.Printf("Uptime: %v", localMgr.GetRoutineUptime(routine.GetID()))
    log.Printf("Done: %v", localMgr.IsRoutineDone(routine.GetID()))
}
```

### Metrics Integration

Enable and configure metrics for observability.

```go
// Enable metrics with URL
globalMgr.UpdateMetadata("SET_METRICS_URL", ":9090")

// Or enable with custom interval
globalMgr.UpdateMetadata("SET_METRICS_URL", []interface{}{
    true,              // enabled
    ":9090",           // URL
    10 * time.Second,  // update interval
})

// Start metrics collector (if not using URL)
metrics.StartCollector()

// Get metrics handler for your HTTP server
mux.Handle("/metrics", metrics.GetMetricsHandler())
```

---

## Best Practices

1. **Always Initialize Global Manager First** - Sets up signal handling and global context
2. **Use Descriptive Names** - Use meaningful app and local names for better organization
3. **Check Context in Loops** - Always check `ctx.Done()` in worker function loops
4. **Enable Panic Recovery** - Keep panic recovery enabled in production (default)
5. **Use Function Wait Groups** - Coordinate multiple goroutines of the same function
6. **Configure Timeouts** - Set appropriate timeouts for long-running operations
7. **Monitor Metrics** - Enable metrics and monitor goroutine counts and health
8. **Graceful Shutdown** - Use safe shutdown (`safe=true`) for graceful termination
9. **Error Handling** - Always check errors from initialization and operations
10. **Organize Hierarchically** - Use the hierarchy to logically group related goroutines

---

## Troubleshooting

### Issue: Goroutines Not Shutting Down

**Solution:** Ensure worker functions check `ctx.Done()` in loops and return when cancelled.

### Issue: Panic Crashes Application

**Solution:** Ensure panic recovery is enabled (default). Check `WithPanicRecovery(true)` is set.

### Issue: Shutdown Timeout

**Solution:** Increase shutdown timeout or investigate why goroutines are stuck. Check for blocking operations that don't respect context cancellation.

### Issue: Context Not Cancelling

**Solution:** Ensure you're using the context passed to the worker function, not creating new contexts.

### Issue: Metrics Not Appearing

**Solution:** Ensure metrics are enabled via `UpdateMetadata("SET_METRICS_URL", ...)` and metrics collector is started.

---

For more information, see the [README.md](README.md) for architecture details and the [metrics documentation](metrics/README.md) for metrics setup.

