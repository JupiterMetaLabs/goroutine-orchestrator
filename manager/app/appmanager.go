package app

import (
	"time"

	LocalHelper "github.com/JupiterMetaLabs/goroutine-orchestrator/internal/helper/local"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/errors"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/interfaces"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/local"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/metrics"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/types"
)

// AppManagerStruct represents an application-level manager that coordinates all local managers
// within a specific application or service. It provides methods to create, manage, and shutdown
// local managers and track goroutines at the application level.
type AppManagerStruct struct {
	// AppName is the unique identifier for this application manager
	AppName string
}

// NewAppManager creates and returns a new AppManager instance for the specified application name.
// This constructor does not initialize the app manager - call CreateApp() to register it with the global manager.
//
// Parameters:
//   - Appname: Unique identifier for the application (e.g., "api-server", "worker-pool")
//
// Returns:
//   - An AppGoroutineManagerInterface implementation that can be used to create and manage local managers
//
// Example:
//
//	appMgr := app.NewAppManager("my-app")
//	app, err := appMgr.CreateApp()
func NewAppManager(Appname string) interfaces.AppGoroutineManagerInterface {
	return &AppManagerStruct{
		AppName: Appname,
	}
}

// CreateApp initializes and registers the app manager with the global manager.
// This method is idempotent - calling it multiple times returns the existing app manager.
// If the global manager is not initialized, it will be created automatically.
//
// The method performs the following operations:
//   - Checks if app manager already exists (returns existing if found)
//   - Auto-initializes global manager if not present
//   - Creates app-level context derived from global context
//   - Registers the app manager with the global manager
//   - Records metrics for the create operation
//
// Returns:
//   - *types.AppManager: The initialized app manager instance
//   - error: nil on success, error if initialization fails
//
// Example:
//
//	appMgr := app.NewAppManager("api-server")
//	app, err := appMgr.CreateApp()
//	if err != nil {
//	    log.Fatalf("Failed to create app: %v", err)
//	}
func (AM *AppManagerStruct) CreateApp() (*types.AppManager, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metrics.RecordManagerOperationDuration("app", "create", duration, AM.AppName)
	}()

	// First check if the app manager is already initialized
	if !types.IsIntilized().App(AM.AppName) {
		// If Global Manager is Not Intilized, then we need to initialize it
		if !types.IsIntilized().Global() {
			global := types.NewGlobalManager().SetGlobalMutex().SetGlobalWaitGroup().SetGlobalContext()
			types.SetGlobalManager(global)
		}
	}

	if types.IsIntilized().App(AM.AppName) {
		return types.GetAppManager(AM.AppName)
	}

	app := types.NewAppManager(AM.AppName).SetAppContext().SetAppMutex()
	types.SetAppManager(AM.AppName, app)

	// Record operation
	metrics.RecordManagerOperation("app", "create", AM.AppName)

	return app, nil
}

// Shutdown gracefully or forcefully shuts down all local managers within this app manager.
// This method coordinates the shutdown of all local managers and their goroutines.
//
// Parameters:
//   - safe: If true, performs graceful shutdown (waits for goroutines with timeout, then force cancels).
//     If false, immediately cancels all contexts without waiting.
//
// Shutdown behavior:
//
//	Safe shutdown (safe=true):
//	- Triggers shutdown on all local managers concurrently
//	- Waits for local managers to complete using wait groups
//	- Each local manager performs graceful → timeout → force cancellation
//	- Records shutdown duration and any errors
//
//	Unsafe shutdown (safe=false):
//	- Immediately cancels all local manager contexts
//	- Cancels app manager's context
//	- No waiting for goroutines to complete
//
// Returns:
//   - error: nil on success, error if app manager not found or shutdown fails
//
// Example:
//
//	Graceful shutdown
//	if err := appMgr.Shutdown(true); err != nil {
//	    log.Printf("Shutdown error: %v", err)
//	}
func (AM *AppManagerStruct) Shutdown(safe bool) error {
	startTime := time.Now()
	shutdownType := "unsafe"
	if safe {
		shutdownType = "safe"
	}

	defer func() {
		duration := time.Since(startTime)
		metrics.RecordShutdownDuration("app", shutdownType, duration, AM.AppName, "")
	}()

	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		metrics.RecordOperationError("manager", "shutdown", "get_app_manager_failed")
		return err
	}

	// Get all local managers
	localManagers, err := AM.GetAllLocalManagers()
	if err != nil {
		metrics.RecordOperationError("manager", "shutdown", "get_local_managers_failed")
		return err
	}

	// Record shutdown operation
	metrics.RecordManagerOperation("app", "shutdown", AM.AppName)

	if safe {
		// Safe shutdown: trigger shutdown on all local managers and wait
		if appManager.Wg != nil {
			// Add all local managers to the wait group
			for _, localMgr := range localManagers {
				appManager.Wg.Add(1)
				go func(lm *types.LocalManager) {
					defer appManager.Wg.Done()

					// Create a LocalManager instance to call Shutdown
					lmInstance := local.NewLocalManager(AM.AppName, lm.LocalName)

					// Call Shutdown on the local manager
					// This will trigger the improved safe shutdown logic (graceful -> timeout -> force)
					_ = lmInstance.Shutdown(true)

					// Wait for local manager's wait group (redundant but safe)
					if lm.Wg != nil {
						lm.Wg.Wait()
					}
				}(localMgr)
			}
			// Wait for all local managers to shutdown
			appManager.Wg.Wait()
		}
	} else {
		// Unsafe shutdown: cancel all local manager contexts forcefully
		for _, localMgr := range localManagers {
			// Create a LocalManager instance to call Shutdown
			lmInstance := local.NewLocalManager(AM.AppName, localMgr.LocalName)

			// Call Shutdown(false) which handles cancellation
			_ = lmInstance.Shutdown(false)
		}

		// Cancel the app manager's context
		if appManager.Cancel != nil {
			appManager.Cancel()
		}
	}

	return nil
}

// NewLocalManager creates a new local manager within this app manager.
// A local manager is used to organize and manage goroutines for a specific module or file.
//
// Parameters:
//   - localName: Unique identifier for the local manager within this app (e.g., "handlers", "workers", "jobs")
//
// The method performs the following:
//   - Creates a new LocalManager instance
//   - Initializes and registers it with this app manager
//   - Records metrics for the create operation
//
// Returns:
//   - *interfaces.LocalGoroutineManagerInterface: The initialized local manager instance
//   - error: Returns error if local manager creation fails or if not found
//
// Example:
//
//	localMgr, err := appMgr.NewLocalManager("http-handlers")
//	if err != nil {
//	    log.Fatalf("Failed to create local manager: %v", err)
//	}
func (AM *AppManagerStruct) NewLocalManager(localName string) (*interfaces.LocalGoroutineManagerInterface, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metrics.RecordManagerOperationDuration("local", "create", duration, AM.AppName)
	}()

	// Use the LocalManagerCreator interface to create a new local manager
	localManager := local.NewLocalManager(AM.AppName, localName)
	if localManager == nil {
		metrics.RecordOperationError("manager", "create_local", "local_manager_not_found")
		return nil, errors.ErrLocalManagerNotFound
	}
	_, err := localManager.CreateLocal(localName)
	if err != nil {
		metrics.RecordOperationError("manager", "create_local", "create_failed")
		return nil, err
	}

	return &localManager, nil
}

// GetAllLocalManagers retrieves all local managers registered with this app manager.
// This returns a slice of all local manager instances within the application.
//
// Returns:
//   - []*types.LocalManager: Slice of all local managers in this app
//   - error: Returns error if app manager is not found or not initialized
//
// Note: This creates a new slice from the internal map. For just counting, use GetLocalManagerCount() instead.
//
// Example:
//
//	locals, err := appMgr.GetAllLocalManagers()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	}
//	for _, local := range locals {
//	    fmt.Printf("Local: %s\n", local.LocalName)
//	}
func (AM *AppManagerStruct) GetAllLocalManagers() ([]*types.LocalManager, error) {
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return nil, err
	}
	return LocalHelper.NewLocalHelper().MapToSlice(appManager.GetLocalManagers()), nil
}

// GetLocalManager retrieves a specific local manager by name from this app manager.
//
// Parameters:
//   - localName: The unique name of the local manager to retrieve
//
// Returns:
//   - *types.LocalManager: The requested local manager instance
//   - error: Returns error if app manager or local manager is not found
//
// Example:
//
//	localMgr, err := appMgr.GetLocalManager("handlers")
//	if err != nil {
//	    log.Printf("Local manager not found: %v", err)
//	}
func (AM *AppManagerStruct) GetLocalManager(localName string) (*types.LocalManager, error) {
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return nil, err
	}
	return appManager.GetLocalManager(localName)
}

// GetAllGoroutines retrieves all tracked goroutines across all local managers in this app.
// This aggregates routines from all local managers within the application.
//
// WARNING: This method creates a new slice and should be used sparingly as it can consume
// significant memory if there are many goroutines. For just counting, use GetGoroutineCount() instead.
//
// Returns:
//   - []*types.Routine: Slice containing all goroutines in this app manager
//   - error: Returns error if app manager is not found
//
// Time Complexity: O(n*m) where n is the number of local managers and m is the average number of goroutines per local manager
//
// Example:
//
//	goroutines, err := appMgr.GetAllGoroutines()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	}
//	fmt.Printf("Total goroutines in app: %d\n", len(goroutines))
func (AM *AppManagerStruct) GetAllGoroutines() ([]*types.Routine, error) {
	// Return the All Goroutines for the particular app manager
	// Dont use this unless you need to get all the goroutines for the particular app manager. This would take significant memory.
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return nil, err
	}
	LocalManagers := appManager.GetLocalManagers()
	allGoroutines := make([]*types.Routine, 0)
	for _, localManager := range LocalManagers {
		goroutines := LocalHelper.NewLocalHelper().RoutinesMapToSlice(localManager.GetRoutines())
		allGoroutines = append(allGoroutines, goroutines...)
	}
	return allGoroutines, nil
}

// GetGoroutineCount returns the total number of tracked goroutines in this app manager.
// This method aggregates counts from all local managers without creating intermediate slices,
// making it more memory efficient than GetAllGoroutines().
//
// Returns:
//   - int: Total count of goroutines across all local managers in this app
//     Returns 0 if app manager is not found
//
// Performance: O(n) where n is the number of local managers, but with minimal memory allocation
//
// Example:
//
//	count := appMgr.GetGoroutineCount()
//	log.Printf("App has %d active goroutines", count)
func (AM *AppManagerStruct) GetGoroutineCount() int {
	// Dont Use GetAllGoroutines() as it will create a new slice - memory usage would be O(n)
	// and it will be a performance issue
	// Return the Go Routine count for the particular app manager
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return 0
	}
	LocalManagers := appManager.GetLocalManagers()
	count := 0
	for _, localManager := range LocalManagers {
		count += localManager.GetRoutineCount()
	}
	return count
}

// GetLocalManagerCount returns the total number of local managers registered with this app manager.
//
// Returns:
//   - int: Count of local managers in this app
//     Returns 0 if app manager is not found
//
// Example:
//
//	count := appMgr.GetLocalManagerCount()
//	log.Printf("App has %d local managers", count)
func (AM *AppManagerStruct) GetLocalManagerCount() int {
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return 0
	}
	return appManager.GetLocalManagerCount()
}

// GetLocalManagerByName retrieves a specific local manager by its name.
// This is an alias for GetLocalManager() provided for API consistency.
//
// Parameters:
//   - localName: The unique name of the local manager to retrieve
//
// Returns:
//   - *types.LocalManager: The requested local manager instance
//   - error: Returns error if app manager or local manager is not found
//
// Example:
//
//	localMgr, err := appMgr.GetLocalManagerByName("workers")
//	if err != nil {
//	    log.Printf("Local manager not found: %v", err)
//	}
func (AM *AppManagerStruct) GetLocalManagerByName(localName string) (*types.LocalManager, error) {
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return nil, err
	}
	return appManager.GetLocalManager(localName)
}

// Get retrieves a specific app manager by its name.
//
// Returns:
//   - *types.AppManager: The requested app manager instance
//   - error: Returns error if app manager is not found
//
// Example:
//
//	appMgr, err := appMgr.Get()
//	if err != nil {
//	    log.Printf("App manager not found: %v", err)
//	}
func (AM *AppManagerStruct) Get() (*types.AppManager, error) {
	return types.GetAppManager(AM.AppName)
}
