package global

import (
	"time"

	AppHelper "github.com/JupiterMetaLabs/goroutine-orchestrator/internal/helper/app"
	LocalHelper "github.com/JupiterMetaLabs/goroutine-orchestrator/internal/helper/local"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/app"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/interfaces"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/metrics"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/types"
)

// GlobalManagerStruct is the singleton manager that orchestrates the entire application lifecycle.
// It manages all app-level managers, provides process-wide context with signal handling,
// and coordinates global shutdown operations. This is the top-level component in the
// hierarchical manager system (Global → App → Local → Routine).
type GlobalManagerStruct struct{}

// NewGlobalManager creates and returns a new GlobalManager instance.
// This is a lightweight constructor that only creates the struct - call Init() to
// initialize the global manager and set up signal handling.
//
// Returns:
//   - An implementation of GlobalGoroutineManagerInterface
//
// Example:
//
//	globalMgr := global.NewGlobalManager()
//	_, err := globalMgr.Init()
func NewGlobalManager() interfaces.GlobalGoroutineManagerInterface {
	return &GlobalManagerStruct{}
}

// Init initializes the global manager and sets up signal handling for graceful shutdown.
// This method is idempotent - calling it multiple times returns the existing global manager.
//
// The method performs the following operations:
//   - Checks if global manager already exists (returns existing if found)
//   - Creates global context with signal handling (SIGINT, SIGTERM)
//   - Initializes global mutex for thread-safe operations
//   - Sets up global wait group for coordinated shutdown
//   - Initializes metadata with default values
//   - Records metrics for the init operation
//
// Signal Handling:
//
//	The global context automatically listens for SIGINT (Ctrl+C) and SIGTERM signals
//	and triggers graceful shutdown when received.
//
// Returns:
//   - *types.GlobalManager: The initialized global manager instance
//   - error: nil on success, error if initialization fails
//
// Example:
//
//	globalMgr := global.NewGlobalManager()
//	_, err := globalMgr.Init()
//	if err != nil {
//	    log.Fatalf("Failed to initialize global manager: %v", err)
//	}
func (GM *GlobalManagerStruct) Init() (*types.GlobalManager, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metrics.RecordManagerOperationDuration("global", "init", duration, "")
	}()

	if types.IsIntilized().Global() {
		return types.GetGlobalManager()
	}

	Global := types.NewGlobalManager().SetGlobalMutex().SetGlobalWaitGroup().SetGlobalContext()
	types.SetGlobalManager(Global)

	// Record operation
	metrics.RecordManagerOperation("global", "init", "")

	return types.GetGlobalManager()
}

// Shutdown gracefully or forcefully shuts down all app managers and the global context.
// This is the top-level shutdown method that coordinates shutdown across the entire application.
//
// Parameters:
//   - safe: If true, performs graceful shutdown (waits for goroutines with timeout, then force cancels).
//     If false, immediately cancels all contexts without waiting.
//
// Shutdown Flow:
//
//	Safe shutdown (safe=true):
//	- Iterates through all app managers concurrently
//	- Each app manager triggers shutdown on all local managers
//	- Waits for all app managers to complete using wait groups
//	- Each level performs graceful → timeout → force cancellation
//	- Records shutdown duration and metrics
//
//	Unsafe shutdown (safe=false):
//	- Immediately triggers shutdown(false) on all app managers
//	- Cancels the global context
//	- No waiting for goroutines to complete
//	- Risk of data loss or incomplete cleanup
//
// Returns:
//   - error: nil on success, error if global manager not found or shutdown fails
//
// Note: This is typically called automatically by signal handlers (SIGINT/SIGTERM).
// Manual invocation is only needed for programmatic shutdown.
//
// Example:
//
//	// Graceful shutdown
//	if err := globalMgr.Shutdown(true); err != nil {
//	    log.Printf("Shutdown error: %v", err)
//	}
func (GM *GlobalManagerStruct) Shutdown(safe bool) error {
	startTime := time.Now()
	shutdownType := "unsafe"
	if safe {
		shutdownType = "safe"
	}

	defer func() {
		duration := time.Since(startTime)
		metrics.RecordShutdownDuration("global", shutdownType, duration, "", "")
	}()

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		metrics.RecordOperationError("manager", "shutdown", "get_global_manager_failed")
		return err
	}

	// Get all app managers
	appManagers, err := GM.GetAllAppManagers()
	if err != nil {
		metrics.RecordOperationError("manager", "shutdown", "get_app_managers_failed")
		return err
	}

	// Record shutdown operation
	metrics.RecordManagerOperation("global", "shutdown", "")

	if safe {
		// Safe shutdown: trigger shutdown on all app managers and wait
		if globalMgr.Wg != nil {
			// Add all app managers to the wait group
			for _, appMgr := range appManagers {
				globalMgr.Wg.Add(1)
				go func(am *types.AppManager) {
					defer globalMgr.Wg.Done()

					// Create an AppManager instance to call Shutdown
					amInstance := app.NewAppManager(am.AppName)

					// Call Shutdown on the app manager
					// This will trigger AppManager.Shutdown -> LocalManager.Shutdown
					_ = amInstance.Shutdown(true)

					// Wait for app manager's wait group (redundant but safe)
					// Lock to safely read Wg pointer to avoid race condition
					am.LockAppReadMutex()
					wg := am.Wg
					am.UnlockAppReadMutex()
					if wg != nil {
						wg.Wait()
					}
				}(appMgr)
			}
			// Wait for all app managers to shutdown
			globalMgr.Wg.Wait()
		}
	} else {
		// Unsafe shutdown: cancel all app manager contexts forcefully
		for _, appMgr := range appManagers {
			// Create an AppManager instance to call Shutdown
			amInstance := app.NewAppManager(appMgr.AppName)

			// Call Shutdown(false) which handles cancellation
			_ = amInstance.Shutdown(false)
		}

		// Cancel the global manager's context
		if globalMgr.Cancel != nil {
			globalMgr.Cancel()
		}
	}

	return nil
}

// GetAllAppManagers retrieves all app managers registered with the global manager.
// This returns a slice of all app manager instances across the entire application.
//
// Returns:
//   - []*types.AppManager: Slice of all app managers
//   - error: Returns error if global manager is not initialized
//
// Note: This creates a new slice from the internal map. For just counting, use GetAppManagerCount() instead.
//
// Example:
//
//	apps, err := globalMgr.GetAllAppManagers()
//	if err != nil {
//	    log.Printf("Error getting app managers: %v", err)
//	}
//	for _, app := range apps {
//	    fmt.Printf("App: %s\n", app.AppName)
//	}
func (GM *GlobalManagerStruct) GetAllAppManagers() ([]*types.AppManager, error) {
	Global, err := types.GetGlobalManager()
	if err != nil {
		return nil, err
	}

	mapValue := Global.GetAppManagers()
	helper := AppHelper.NewAppHelper()
	return helper.MapToSlice(mapValue), nil
}

// GetAppManagerCount returns the total number of app managers in the global manager.
//
// Returns:
//   - int: Count of app managers
//     Returns 0 if global manager is not initialized
//
// Example:
//
//	count := globalMgr.GetAppManagerCount()
//	log.Printf("Total app managers: %d", count)
func (GM *GlobalManagerStruct) GetAppManagerCount() int {
	Global, err := types.GetGlobalManager()
	if err != nil {
		return 0
	}
	return Global.GetAppManagerCount()
}

// GetAllLocalManagers retrieves all local managers across all app managers in the system.
// This aggregates local managers from every app manager registered with the global manager.
//
// Returns:
//   - []*types.LocalManager: Slice of all local managers across all apps
//   - error: Returns error if global manager is not initialized
//
// Note: This creates a new slice and should be used sparingly. For just counting,
// use GetLocalManagerCount() instead.
//
// Example:
//
//	locals, err := globalMgr.GetAllLocalManagers()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	}
//	fmt.Printf("Total local managers: %d\n", len(locals))
func (GM *GlobalManagerStruct) GetAllLocalManagers() ([]*types.LocalManager, error) {
	// Get all app managers first
	appManagers, err := GM.GetAllAppManagers()
	if err != nil {
		return nil, err
	}

	// Get all local managers from each app manager
	var localManagers []*types.LocalManager
	for _, appManager := range appManagers {
		// Convert map to slice
		LocalManagerSlice := LocalHelper.NewLocalHelper().MapToSlice(appManager.LocalManagers)
		localManagers = append(localManagers, LocalManagerSlice...)
	}

	return localManagers, nil
}

// GetLocalManagerCount returns the total number of local managers across all apps.
// This method aggregates counts from all app managers without creating intermediate slices,
// making it more memory efficient than GetAllLocalManagers().
//
// Returns:
//   - int: Total count of local managers across all app managers
//     Returns 0 if global manager is not initialized
//
// Performance: O(n) where n is the number of app managers, with minimal memory allocation
//
// Example:
//
//	count := globalMgr.GetLocalManagerCount()
//	log.Printf("Total local managers: %d", count)
func (GM *GlobalManagerStruct) GetLocalManagerCount() int {
	// get all the local managers first
	// Dont use GetAllLocalManagers() as it will create a new slice - memory usage would be O(n)
	// and it will be a performance issue
	App, err := GM.GetAllAppManagers()
	if err != nil {
		return 0
	}
	i := 0
	for _, Apps := range App {
		i += Apps.GetLocalManagerCount()
	}
	return i
}

// GetAllGoroutines retrieves all tracked goroutines across the entire application.
// This aggregates routines from all local managers in all app managers.
//
// WARNING: This method creates a new slice and should be used very sparingly as it can
// consume significant memory if there are many goroutines. For just counting, use
// GetGoroutineCount() instead.
//
// Returns:
//   - []*types.Routine: Slice containing all goroutines in the system
//   - error: Returns error if global manager is not initialized
//
// Time Complexity: O(n*m*k) where:
//   - n = number of app managers
//   - m = average number of local managers per app
//   - k = average number of goroutines per local manager
//
// Example:
//
//	goroutines, err := globalMgr.GetAllGoroutines()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	}
//	fmt.Printf("Total goroutines in system: %d\n", len(goroutines))
func (GM *GlobalManagerStruct) GetAllGoroutines() ([]*types.Routine, error) {
	// Get all app managers first
	appManagers, err := GM.GetAllAppManagers()
	if err != nil {
		return nil, err
	}

	// Get all goroutines from each app manager - would run on O(n*m)
	var goroutines []*types.Routine
	for _, appManager := range appManagers {
		LocalManagers := appManager.GetLocalManagers()
		for _, localManager := range LocalManagers {
			Routines := localManager.GetRoutines()
			goroutines = append(goroutines, LocalHelper.NewLocalHelper().RoutinesMapToSlice(Routines)...)
		}
	}
	return goroutines, nil
}

// GetGoroutineCount returns the total number of tracked goroutines across the entire system.
// This method aggregates counts from all local managers in all apps without creating
// intermediate slices, making it highly memory efficient.
//
// Returns:
//   - int: Total count of goroutines across all app and local managers
//     Returns 0 if global manager is not initialized
//
// Performance: O(n*m) where n is the number of apps and m is average local managers per app,
// but with minimal memory allocation (only counters).
//
// Example:
//
//	count := globalMgr.GetGoroutineCount()
//	log.Printf("Total active goroutines: %d", count)
func (GM *GlobalManagerStruct) GetGoroutineCount() int {
	// get all the goroutines first
	// Dont use GetAllGoroutines() as it will create a new slice - memory usage would be O(n)
	// and it will be a performance issue
	App, err := GM.GetAllAppManagers()
	if err != nil {
		return 0
	}
	i := 0
	for _, Apps := range App {
		LocalManagers := Apps.GetLocalManagers()
		for _, LocalManager := range LocalManagers {
			i += LocalManager.GetRoutineCount()
		}
	}
	return i
}

// UpdateMetadata updates global configuration metadata using the specified flag and value.
// This method allows runtime configuration of shutdown timeouts, metrics settings,
// goroutine limits, and other global parameters.
//
// Parameters:
//   - flag: Configuration flag (e.g., "SET_SHUTDOWN_TIMEOUT", "SET_METRICS_URL", "SET_MAX_ROUTINES")
//   - value: Configuration value (type depends on flag)
//
// Available Flags:
//   - SET_SHUTDOWN_TIMEOUT: time.Duration - graceful shutdown timeout
//   - SET_METRICS_URL: string or []interface{} - enable metrics and set endpoint
//   - SET_MAX_ROUTINES: int - maximum routines limit
//   - SET_UPDATE_INTERVAL: time.Duration - metrics update interval
//
// Returns:
//   - *types.Metadata: Updated metadata configuration
//   - error: Returns error if flag is invalid or value type is incorrect
//
// Example:
//
//	metadata, err := globalMgr.UpdateMetadata("SET_SHUTDOWN_TIMEOUT", 30*time.Second)
//	if err != nil {
//	    log.Printf("Failed to update metadata: %v", err)
//	}
func (GM *GlobalManagerStruct) UpdateMetadata(flag string, value interface{}) (*types.Metadata, error) {
	return GM.UpdateGlobalMetadata(flag, value)
}

// GetMetadata retrieves the current global metadata configuration.
// This returns the metadata instance containing all configuration settings like
// shutdown timeout, metrics enablement, max routines, and update intervals.
//
// Returns:
//   - *types.Metadata: Current metadata configuration
//   - error: Returns error if global manager is not initialized
//
// Example:
//
//	metadata, err := globalMgr.GetMetadata()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	}
//	if metadata.GetMetrics() {
//	    log.Println("Metrics are enabled")
//	}
func (GM *GlobalManagerStruct) GetMetadata() (*types.Metadata, error) {
	return GM.GetGlobalMetadata()
}
