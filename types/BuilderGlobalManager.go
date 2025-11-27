package types

import (
	"context"
	"sync"

	"github.com/neerajchowdary889/GoRoutinesManager/Context"
	"github.com/neerajchowdary889/GoRoutinesManager/types/Errors"
)

func NewGlobalManager() *GlobalManager {
	if IsIntilized().Global() {
		return Global
	}
	GlobalManger := &GlobalManager{
		AppManagers: make(map[string]*AppManager),
	}

	return GlobalManger
}

// Mutex Lock APIs
// LockGlobalReadMutex locks the global read mutex for the global manager - This is used to read the global manager's data
func (GM *GlobalManager) LockGlobalReadMutex() {
	if GM.GlobalMu == nil {
		GM.SetGlobalMutex()
	}
	GM.GlobalMu.RLock()
}

// UnlockGlobalReadMutex unlocks the global read mutex for the global manager - This is used to read the global manager's data
func (GM *GlobalManager) UnlockGlobalReadMutex() {
	if GM.GlobalMu == nil {
		GM.SetGlobalMutex()
	}
	GM.GlobalMu.RUnlock()
}

// LockGlobalWriteMutex locks the global write mutex for the global manager - This is used to write the global manager's data
func (GM *GlobalManager) LockGlobalWriteMutex() {
	if GM.GlobalMu == nil {
		GM.SetGlobalMutex()
	}
	GM.GlobalMu.Lock()
}

// UnlockGlobalWriteMutex unlocks the global write mutex for the global manager - This is used to write the global manager's data
func (GM *GlobalManager) UnlockGlobalWriteMutex() {
	if GM.GlobalMu == nil {
		GM.SetGlobalMutex()
	}
	GM.GlobalMu.Unlock()
}

// >>> Set APIs
// SetGlobalMutex sets the global mutex for the global manager
func (GM *GlobalManager) SetGlobalMutex() *GlobalManager {
	GM.GlobalMu = &sync.RWMutex{}
	return GM
}

// SetGlobalContext sets the global context and cancel function for the global manager
func (GM *GlobalManager) SetGlobalContext() *GlobalManager {
	GM.Ctx = Context.GetGlobalContext().Get()
	GM.Cancel = func() {
		Context.GetGlobalContext().Done(GM.Ctx)
	}
	return GM
}

// SetGlobalWaitGroup sets the global wait group for the global manager - This is used to concurrently wait for all app managers to shutdown
func (GM *GlobalManager) SetGlobalWaitGroup() *GlobalManager {
	GM.Wg = &sync.WaitGroup{}
	return GM
}

// AddAppManager adds a new app manager to the global manager
func (GM *GlobalManager) AddAppManager(appName string, app *AppManager) *GlobalManager {
	if IsIntilized().App(appName) {
		return GM
	}
	GM.LockGlobalWriteMutex()
	defer GM.UnlockGlobalWriteMutex()
	GM.AppManagers[appName] = app
	return GM
}

// RemoveAppManager removes an app manager from the global manager
func (GM *GlobalManager) RemoveAppManager(appName string) *GlobalManager {
	GM.LockGlobalWriteMutex()
	defer GM.UnlockGlobalWriteMutex()
	delete(GM.AppManagers, appName)
	return GM
}

// >>> Get APIs
// GetGlobalMutex gets the global mutex for the global manager
func (GM *GlobalManager) GetGlobalMutex() *sync.RWMutex {
	return GM.GlobalMu
}

// GetGlobalContext gets the global context for the global manager
func (GM *GlobalManager) GetGlobalContext() (context.Context, context.CancelFunc) {
	return GM.Ctx, GM.Cancel
}

// GetGlobalWaitGroup gets the global wait group for the global manager
func (GM *GlobalManager) GetGlobalWaitGroup() *sync.WaitGroup {
	return GM.Wg
}

// GetAppManagers gets all the app managers for the global manager
func (GM *GlobalManager) GetAppManagers() map[string]*AppManager {
	GM.LockGlobalReadMutex()
	defer GM.UnlockGlobalReadMutex()
	return GM.AppManagers
}

// GetAppManager gets a specific app manager for the global manager
func (GM *GlobalManager) GetAppManager(appName string) (*AppManager, error) {
	if !IsIntilized().App(appName) {
		return nil, Errors.ErrAppManagerNotFound
	}
	GM.LockGlobalReadMutex()
	defer GM.UnlockGlobalReadMutex()
	return GM.AppManagers[appName], nil
}

// GetAppManagerCount gets the number of app managers for the global manager
func (GM *GlobalManager) GetAppManagerCount() int {
	GM.LockGlobalReadMutex()
	defer GM.UnlockGlobalReadMutex()
	return len(GM.AppManagers)
}
