package types

import (
	"sync"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Errors"
)

var (
	once sync.Once
	lock sync.RWMutex
)

func SetGlobalManager(global *GlobalManager) {
	// By using this once - we can avoid the race condition thus made thread safe
	once.Do(func() {
		Global = global
	})
}

func SetAppManager(appName string, app *AppManager) {
	if IsIntilized().App(appName) {
		return
	}
	Global.AddAppManager(appName, app)
}

func SetLocalManager(appName, localName string, local *LocalManager) {
	if IsIntilized().Local(appName, localName) {
		return
	}
	// Get the appmanager first
	appManager, err := Global.GetAppManager(appName)
	if err != nil {
		return
	}
	appManager.AddLocalManager(localName, local)
}

func GetGlobalManager() (*GlobalManager, error) {
	// Read Lock and Unlock after returning the global manager
	lock.RLock()
	defer lock.RUnlock()

	if Global == nil {
		return nil, Errors.ErrGlobalManagerNotFound
	}
	return Global, nil
}

func GetAppManager(appName string) (*AppManager, error) {
	if !IsIntilized().App(appName) {
		return nil, Errors.ErrAppManagerNotFound
	}
	return Global.GetAppManager(appName)
}

func GetLocalManager(appName, localName string) (*LocalManager, error) {
	if !IsIntilized().App(appName) {
		return nil, Errors.ErrAppManagerNotFound
	}
	appManager, err := Global.GetAppManager(appName)
	if err != nil {
		return nil, err
	}
	return appManager.GetLocalManager(localName)
}
