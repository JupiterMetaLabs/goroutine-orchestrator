package types

type Initializer struct{}

func IsIntilized() Initializer {
	return Initializer{}
}

func (Is Initializer) Global() bool {
	return Global != nil
}

// Made everything thread safe
// Check if the app manager is alread intilized in the global manager
func (Is Initializer) App(appName string) bool {
	if Global == nil {
		return false
	}
	// RLock and RUnlock
	Global.LockGlobalReadMutex()
	_, ok := Global.AppManagers[appName]
	Global.UnlockGlobalReadMutex()
	return ok
}

// Thread safe check if the local manager is alread intilized in the app manager
func (Is Initializer) Local(appName, localName string) bool {
	if Global == nil || !Is.App(appName) {
		return false
	}

	// Global RLock and RUnlock
	Global.LockGlobalReadMutex()
	appMgr, ok := Global.AppManagers[appName]	
	Global.UnlockGlobalReadMutex()
	if !ok {
		return false
	}
	
	// App RLock and RUnlock
	appMgr.LockAppReadMutex()
	_, ok = appMgr.LocalManagers[localName]
	appMgr.UnlockAppReadMutex()
	return ok
}

// Thread safe check if the routine is alread intilized in the local manager
func (Is Initializer) Routine(appName, localName, routineID string) bool {
	if Global == nil || !Is.App(appName) {
		return false
	}
	
	// Global RLock and RUnlock
	Global.LockGlobalReadMutex()
	appMgr, ok := Global.AppManagers[appName]	
	Global.UnlockGlobalReadMutex()
	if !ok {
		return false
	}
	
	// App RLock and RUnlock
	appMgr.LockAppReadMutex()
	localMgr, ok := appMgr.LocalManagers[localName]
	appMgr.UnlockAppReadMutex()
	if !ok {
		return false
	}
	
	// Local RLock and RUnlock
	localMgr.LockLocalReadMutex()
	_, ok = localMgr.Routines[routineID]
	localMgr.UnlockLocalReadMutex()
	return ok
}
