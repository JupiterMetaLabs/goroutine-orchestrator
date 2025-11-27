package App

import (
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

type AppManager struct{
	AppName string
}

func NewAppManager(Appname string) Interface.AppGoroutineManagerInterface {
	return &AppManager{
		AppName: Appname,
	}
}

func (AM *AppManager) CreateApp(appName string) (*types.AppManager, error) {
	// First check if the app manager is already initialized
	if !types.IsIntilized().App(appName) {
		// If Global Manager is Not Intilized, then we need to initialize it
		globalManager := Global.NewGlobalManager()
		err := globalManager.Init()
		if err != nil {
			return nil, err
		}
	}

	if types.IsIntilized().App(appName) {
		return types.GetAppManager(appName)
	}

	app := types.NewAppManager(appName).SetAppContext().SetAppMutex()
	types.SetAppManager(appName, app)

	return app, nil
}

func (AM *AppManager) Shutdown(safe bool) error {
	//TODO
	return nil
}

func (AM *AppManager) CreateLocal(localName string) (*types.LocalManager, error) {

	return nil, nil
}

func (AM *AppManager) GetAllLocalManagers() ([]*types.LocalManager, error) {
	return nil, nil
}

func (AM *AppManager) GetLocalManager(localName string) (*types.LocalManager, error) {
	return nil, nil
}

func (AM *AppManager) GetAllGoroutines() ([]*types.Routine, error) {
	// TODO: Implement logic to collect all goroutines from all local managers
	return nil, nil
}

func (AM *AppManager) GetGoroutineCount() int {
	// TODO: Implement logic to count all goroutines across all local managers
	return 0
}

func (AM *AppManager) GetLocalManagerCount() int {
	// TODO: Implement logic to count all local managers
	return 0
}
