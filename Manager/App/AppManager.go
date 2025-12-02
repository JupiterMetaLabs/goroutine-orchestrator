package App

import (
	"time"

	LocalHelper "github.com/neerajchowdary889/GoRoutinesManager/Helper/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/metrics"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Errors"
)

type AppManager struct {
	AppName string
}

func NewAppManager(Appname string) Interface.AppGoroutineManagerInterface {
	return &AppManager{
		AppName: Appname,
	}
}

func (AM *AppManager) CreateApp() (*types.AppManager, error) {
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

func (AM *AppManager) Shutdown(safe bool) error {
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
					lmInstance := Local.NewLocalManager(AM.AppName, lm.LocalName)

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
			lmInstance := Local.NewLocalManager(AM.AppName, localMgr.LocalName)

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

func (AM *AppManager) CreateLocal(localName string) (*types.LocalManager, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metrics.RecordManagerOperationDuration("local", "create", duration, AM.AppName)
	}()

	// Use the LocalManagerCreator interface to create a new local manager
	localManager := Local.NewLocalManager(AM.AppName, localName)
	if localManager == nil {
		metrics.RecordOperationError("manager", "create_local", "local_manager_not_found")
		return nil, Errors.ErrLocalManagerNotFound
	}
	Manager, err := localManager.CreateLocal(localName)
	if err != nil {
		metrics.RecordOperationError("manager", "create_local", "create_failed")
		return nil, err
	}

	// Record operation
	metrics.RecordManagerOperation("local", "create", AM.AppName)

	return Manager, nil
}

func (AM *AppManager) GetAllLocalManagers() ([]*types.LocalManager, error) {
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return nil, err
	}
	return LocalHelper.NewLocalHelper().MapToSlice(appManager.GetLocalManagers()), nil
}

func (AM *AppManager) GetLocalManager(localName string) (*types.LocalManager, error) {
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return nil, err
	}
	return appManager.GetLocalManager(localName)
}

func (AM *AppManager) GetAllGoroutines() ([]*types.Routine, error) {
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

func (AM *AppManager) GetGoroutineCount() int {
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

func (AM *AppManager) GetLocalManagerCount() int {
	// Return the Local Manager count for the particular app manager
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return 0
	}
	return appManager.GetLocalManagerCount()
}

func (AM *AppManager) GetLocalManagerByName(localName string) (*types.LocalManager, error) {
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return nil, err
	}
	return appManager.GetLocalManager(localName)
}