package Global

import (
	"time"

	AppHelper "github.com/neerajchowdary889/GoRoutinesManager/Helper/App"
	LocalHelper "github.com/neerajchowdary889/GoRoutinesManager/Helper/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/App"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
	"github.com/neerajchowdary889/GoRoutinesManager/metrics"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

type GlobalManager struct{}

func NewGlobalManager() Interface.GlobalGoroutineManagerInterface {
	return &GlobalManager{}
}

func (GM *GlobalManager) Init() error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metrics.RecordManagerOperationDuration("global", "init", duration, "")
	}()

	if types.IsIntilized().Global() {
		return nil
	}

	Global := types.NewGlobalManager().SetGlobalMutex().SetGlobalWaitGroup().SetGlobalContext()
	types.SetGlobalManager(Global)

	// Record operation
	metrics.RecordManagerOperation("global", "init", "")

	return nil
}

func (GM *GlobalManager) Shutdown(safe bool) error {
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
					amInstance := App.NewAppManager(am.AppName)

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
			amInstance := App.NewAppManager(appMgr.AppName)

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

func (GM *GlobalManager) GetAllAppManagers() ([]*types.AppManager, error) {
	Global, err := types.GetGlobalManager()
	if err != nil {
		return nil, err
	}

	mapValue := Global.GetAppManagers()
	helper := AppHelper.NewAppHelper()
	return helper.MapToSlice(mapValue), nil
}

func (GM *GlobalManager) GetAppManagerCount() int {
	Global, err := types.GetGlobalManager()
	if err != nil {
		return 0
	}
	return Global.GetAppManagerCount()
}

func (GM *GlobalManager) GetAllLocalManagers() ([]*types.LocalManager, error) {
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

func (GM *GlobalManager) GetLocalManagerCount() int {
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

func (GM *GlobalManager) GetAllGoroutines() ([]*types.Routine, error) {
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

func (GM *GlobalManager) GetGoroutineCount() int {
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

func (GM *GlobalManager) UpdateMetadata(flag string, value interface{}) (*types.Metadata, error) {
	return GM.UpdateGlobalMetadata(flag, value)
}

func (GM *GlobalManager) GetMetadata() (*types.Metadata, error) {
	return GM.GetGlobalMetadata()
}
