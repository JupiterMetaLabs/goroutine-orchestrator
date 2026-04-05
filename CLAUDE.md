# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library providing hierarchical goroutine supervision with lifecycle management, graceful shutdown, and observability. It uses a 4-level hierarchy: GlobalManager → AppManager → LocalManager → Routine.

## Build Commands

```bash
# Build the library
go build ./...

# Run all tests
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run a single test
go test -v -run TestName ./test/manager/...

# Run tests for a specific package
go test -v ./manager/local/...

# Lint (requires golangci-lint)
golangci-lint run --timeout=5m

# Vet
go vet ./...

# Format code
go fmt ./...

# Verify dependencies
go mod verify
```

## Architecture

### Manager Hierarchy

The library uses a 4-level hierarchical architecture:

1. **GlobalManager** (`manager/global/`, `types/types.go`): Singleton process-level manager
   - Manages all AppManagers
   - Sets up signal handling (SIGINT, SIGTERM)
   - Coordinates global shutdown
   - Access via `global.NewGlobalManager()` → `Init()`

2. **AppManager** (`manager/app/`): Application-level manager
   - Manages LocalManagers for a specific app/service
   - Creates app-level context (child of global)
   - Access via `app.NewAppManager(name)` → `CreateApp()`

3. **LocalManager** (`manager/local/`): Module/file-level manager
   - Spawns and tracks individual goroutines (Routines)
   - Manages function-level wait groups
   - Main method: `Go(functionName, workerFunc, opts...)`
   - Access via `local.NewLocalManager(appName, localName)` → `CreateLocal()`

4. **Routine** (`types/types.go`): Individual tracked goroutine
   - Unique ID, function name, context, done channel
   - Automatically tracked and cleaned up

### Context Flow

- `ctxo/` package provides hierarchical context management
- GlobalContext (process-level) → AppContext → LocalContext → RoutineContext
- Cancellation propagates down the hierarchy automatically
- Signal handling triggers graceful shutdown via context cancellation

### Key Types (types package)

- `GlobalManager`: Contains `AppManagers` map, context, wait group, metadata
- `AppManager`: Contains `LocalManagers` map, app-level context
- `LocalManager`: Contains `Routines` map, `FunctionWgs`, atomic `routineCount`
- `Routine`: ID, function name, context, cancel func, done channel, start time
- `Metadata`: MaxRoutines, Metrics, MetricsURL, UpdateInterval, ShutdownTimeout

### Thread Safety

- All managers use `sync.RWMutex` for protection
- Routine counts use atomic operations for lock-free reads
- Signal handler setup uses `sync.Once`

## Goroutine Spawning Options

Located in `manager/local/options.go`:

- `WithTimeout(duration)`: Auto-cancel goroutine after duration
- `WithPanicRecovery(bool)`: Enable/disable panic recovery (default: true)
- `AddToWaitGroup(functionName)`: Add to function-level wait group

## Testing

Tests are organized under `test/` directory (not alongside source):

- `test/manager/` - Manager tests (global, app, local)
- `test/e2e/` - End-to-end tests
- `test/shutdown/` - Shutdown behavior tests
- `test/waitgroup/` - Wait group tests
- `test/common/` - Test utilities including `ResetGlobalState()`

**Critical for tests**: Call `common.ResetGlobalState()` at the start of each test to reset the singleton GlobalManager state.

## Interfaces

All manager operations are defined in `manager/interfaces/interfaces.go`:

- `GlobalGoroutineManagerInterface`: Full global manager interface
- `AppGoroutineManagerInterface`: App manager interface
- `LocalGoroutineManagerInterface`: Local manager interface (includes `Go()`, `ShutdownFunction()`, etc.)
- `GoroutineOption`: Option type for goroutine configuration

## Metrics

Prometheus metrics integration in `metrics/`:

- 18+ metric types for goroutines, managers, operations, errors
- `metrics.GetMetricsHandler()` returns HTTP handler for integration
- `metrics.StartCollector(interval)` starts periodic collection

## Important Patterns

### Goroutine Spawning
```go
localMgr.Go("function-name", func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // do work
        }
    }
}, local.WithTimeout(5*time.Minute))
```

### Shutdown Strategy
- Safe shutdown: graceful → timeout → force cancellation
- Unsafe shutdown: immediate context cancellation
- Function-level shutdown via `ShutdownFunction(name, timeout)`

### Error Handling
Custom errors in `manager/errors/errors.go`:
- `ErrGlobalManagerNotFound`, `ErrAppManagerNotFound`, `ErrLocalManagerNotFound`
- `WrngLocalManagerAlreadyExists` (warning-level error)

## Configuration

Default timeouts in `types/types.go`:
- `ShutdownTimeout`: 10 seconds
- `UpdateInterval`: 5 seconds

Update via `globalMgr.UpdateMetadata(flag, value)`:
- `SET_SHUTDOWN_TIMEOUT`: time.Duration
- `SET_METRICS_URL`: string or []interface{}
- `SET_MAX_ROUTINES`: int
- `SET_UPDATE_INTERVAL`: time.Duration
