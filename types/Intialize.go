package types


type Initializer struct{}

func IsIntilized() Initializer {
	return Initializer{}
}

func(Is Initializer) Global() bool {
	return Global != nil
}

// Check if the app manager is alread intilized in the global manager
func(Is Initializer) App(appName string) bool {
	return Global.AppManagers[appName] != nil
}

func(Is Initializer) Local(appName, localName string) bool {
	appMgr, ok := Global.AppManagers[appName]
	if !ok || appMgr == nil {
		return false
	}
	return appMgr.LocalManagers[localName] != nil
}

func(Is Initializer) Routine(appName, localName, routineID string) bool {
	appMgr, ok := Global.AppManagers[appName]
	if !ok || appMgr == nil {
		return false
	}
	localMgr, ok := appMgr.LocalManagers[localName]
	if !ok || localMgr == nil {
		return false
	}
	return localMgr.Routines[routineID] != nil
}