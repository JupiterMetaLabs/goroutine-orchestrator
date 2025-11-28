package App

import (
	"github.com/neerajchowdary889/GoRoutinesManager/Helper"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
	"github.com/neerajchowdary889/GoRoutinesManager/types/Errors"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
)

type AppManager struct{
	AppName string
}

func NewAppManager(Appname string) Interface.AppGoroutineManagerInterface {
	return &AppManager{
		AppName: Appname,
	}
}

func (AM *AppManager) CreateApp() (*types.AppManager, error) {
	// First check if the app manager is already initialized
	if !types.IsIntilized().App(AM.AppName) {
		// If Global Manager is Not Intilized, then we need to initialize it
		globalManager := Global.NewGlobalManager()
		err := globalManager.Init()
		if err != nil {
			return nil, err
		}
	}

	if types.IsIntilized().App(AM.AppName) {
		return types.GetAppManager(AM.AppName)
	}

	app := types.NewAppManager(AM.AppName).SetAppContext().SetAppMutex()
	types.SetAppManager(AM.AppName, app)

	return app, nil
}

func (AM *AppManager) Shutdown(safe bool) error {
	//TODO
	return nil
}

func (AM *AppManager) CreateLocal(localName string) (*types.LocalManager, error) {
	// Use the LocalManagerCreator interface to create a new local manager
	localManager := Local.NewLocalManager(AM.AppName, localName)
	if localManager == nil {
		return nil, Errors.ErrLocalManagerNotFound
	}
	Manager, err := localManager.CreateLocal(localName)
	if err != nil {
		return nil, err
	}
	return Manager, nil
}

func (AM *AppManager) GetAllLocalManagers() ([]*types.LocalManager, error) {
	appManager, err := types.GetAppManager(AM.AppName)
	if err != nil {
		return nil, err
	}
	return Helper.NewLocalHelper().MapToSlice(appManager.GetLocalManagers()), nil
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
		goroutines := Helper.NewLocalHelper().RoutinesMapToSlice(localManager.GetRoutines())
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
