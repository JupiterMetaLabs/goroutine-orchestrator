package types
import "github.com/neerajchowdary889/GoRoutinesManager/types/Errors"
func SetGlobalManager(global *GlobalManager) {
	if IsIntilized().Global() {
		return
	}
	Global = global
}

func SetAppManager(appName string, app *AppManager) {
	if IsIntilized().App(appName) {
		return
	}
	Global.AppManagers[appName] = app
}

func SetLocalManager(appName, localName string, local *LocalManager) {
	if IsIntilized().Local(appName, localName) {
		return
	}
	Global.AppManagers[appName].LocalManagers[localName] = local
}

func GetGlobalManager() (*GlobalManager, error) {
		if Global == nil {
			return nil, Errors.ErrGlobalManagerNotFound
	}
	return Global, nil
}

func GetAppManager(appName string) (*AppManager, error) {
	if !IsIntilized().App(appName) {
		return nil, Errors.ErrAppManagerNotFound
	}
	return Global.AppManagers[appName], nil
}

func GetLocalManager(appName, localName string) (*LocalManager, error) {
	if IsIntilized().App(appName) {
		return nil, Errors.ErrAppManagerNotFound
	}
	if _, ok := Global.AppManagers[appName].LocalManagers[localName]; !ok {
		return nil, Errors.ErrLocalManagerNotFound
	}
	return Global.AppManagers[appName].LocalManagers[localName], nil
}