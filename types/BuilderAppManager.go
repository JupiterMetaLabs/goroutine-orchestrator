package types

import (
	"context"
	"fmt"
	"sync"

	"github.com/neerajchowdary889/GoRoutinesManager/Context"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Errors"
)

const (
	Prefix_AppManager = "AppManager."
)

func NewAppManager(appName string) *AppManager {
	if IsIntilized().App(appName) {
		appMgr, err := Global.GetAppManager(appName)
		if err != nil {
			return nil
		}
		return appMgr
	}

	appMgr := &AppManager{
		AppName:       appName,
		LocalManagers: make(map[string]*LocalManager),
		Wg:            &sync.WaitGroup{}, // Initialize wait group for safe shutdown
	}
	appMgr.SetAppContext()

	// Add the app manager to the global manager
	SetAppManager(appName, appMgr)

	return appMgr
}

// Lock APIs
// LockAppReadMutex locks the app read mutex for the app manager - This is used to read the app manager's data
func (AM *AppManager) LockAppReadMutex() {
	if AM.AppMu == nil {
		AM.SetAppMutex()
	}
	AM.AppMu.RLock()
}

// UnlockAppReadMutex unlocks the app read mutex for the app manager - This is used to read the app manager's data
func (AM *AppManager) UnlockAppReadMutex() {
	if AM.AppMu == nil {
		AM.SetAppMutex()
	}
	AM.AppMu.RUnlock()
}

// LockAppWriteMutex locks the app write mutex for the app manager - This is used to write the app manager's data
func (AM *AppManager) LockAppWriteMutex() {
	if AM.AppMu == nil {
		AM.SetAppMutex()
	}
	AM.AppMu.Lock()
}

// UnlockAppWriteMutex unlocks the app write mutex for the app manager - This is used to write the app manager's data
func (AM *AppManager) UnlockAppWriteMutex() {
	if AM.AppMu == nil {
		AM.SetAppMutex()
	}
	AM.AppMu.Unlock()
}

// >>> Set APIs
// SetAppName sets the name of the app for the app manager
func (AM *AppManager) SetAppName(appName string) *AppManager {
	AM.AppName = appName
	return AM
}

// SetAppMutex sets the mutex for the app manager
func (AM *AppManager) SetAppMutex() *AppManager {
	if AM.AppMu == nil {
		AM.AppMu = &sync.RWMutex{}
	}
	return AM
}

// SetAppContext sets the context for the app manager
func (AM *AppManager) SetAppContext() *AppManager {
	ctx := Context.GetAppContext(Prefix_AppManager + AM.AppName).Get()
	Done := func() {
		Context.GetAppContext(Prefix_AppManager + AM.AppName).Done(ctx)
	}
	AM.Ctx = ctx
	AM.Cancel = Done
	return AM
}

// SetAppWaitGroup sets the wait group for the app manager
func (AM *AppManager) SetAppWaitGroup(wg *sync.WaitGroup) *AppManager {
	AM.Wg = wg
	return AM
}

// SetAppParentContext sets the parent context for the app manager
func (AM *AppManager) SetAppParentContext() *AppManager {
	AM.ParentCtx, _ = NewGlobalManager().GetGlobalContext()
	return AM
}

// Create Local Manger function is made private
// Only accessible throught the appmanager
// After creating the local manager, it will be added to the app manager
func (AM *AppManager) CreateLocal(localName string) (*LocalManager, error) {
	// Get or create the local manager
	if IsIntilized().Local(AM.AppName, localName) {
		LM, err := AM.GetLocalManager(localName)
		if err != nil {
			return nil, Errors.ErrLocalManagerNotFound
		}
		return LM, Errors.WrngLocalManagerAlreadyExists
	}
	// Set the parent contex before returning
	LM := newLocalManager(localName, AM.AppName).SetParentContext(AM.ParentCtx)

	AM.AddLocalManager(LM.LocalName, LM)

	return LM, nil
}

// AddLocalManager adds a new local manager to the app manager
func (AM *AppManager) AddLocalManager(localName string, local *LocalManager) *AppManager {
	if IsIntilized().Local(AM.AppName, localName) {
		return AM
	}
	AM.LockAppWriteMutex()
	defer AM.UnlockAppWriteMutex()
	AM.LocalManagers[localName] = local
	return AM
}

// RemoveLocalManager removes a local manager from the app manager
func (AM *AppManager) RemoveLocalManager(localName string) *AppManager {
	AM.LockAppWriteMutex()
	defer AM.UnlockAppWriteMutex()
	delete(AM.LocalManagers, localName)
	return AM
}

// >>> Get APIs
// GetLocalManagers gets all the local managers for the app manager
func (AM *AppManager) GetLocalManagers() map[string]*LocalManager {
	AM.LockAppReadMutex()
	defer AM.UnlockAppReadMutex()
	return AM.LocalManagers
}

// GetLocalManager gets a specific local manager for the app manager
func (AM *AppManager) GetLocalManager(localName string) (*LocalManager, error) {
	AM.LockAppReadMutex()
	defer AM.UnlockAppReadMutex()
	if _, ok := AM.LocalManagers[localName]; !ok {
		return nil, fmt.Errorf("%w: %s", Errors.ErrLocalManagerNotFound, localName)
	}
	return AM.LocalManagers[localName], nil
}

// GetLocalManagerCount gets the number of local managers for the app manager
func (AM *AppManager) GetLocalManagerCount() int {
	AM.LockAppReadMutex()
	defer AM.UnlockAppReadMutex()
	return len(AM.LocalManagers)
}

// GetAppName gets the name of the app for the app manager
func (AM *AppManager) GetAppName() string {
	AM.LockAppReadMutex()
	defer AM.UnlockAppReadMutex()
	return AM.AppName
}

// GetAppContext gets the context for the app manager
func (AM *AppManager) GetAppContext() (context.Context, context.CancelFunc) {
	AM.LockAppReadMutex()
	defer AM.UnlockAppReadMutex()
	return AM.Ctx, AM.Cancel
}

// Get the Parent Context for the app manager
func (AM *AppManager) GetParentContext() context.Context {
	AM.LockAppReadMutex()
	defer AM.UnlockAppReadMutex()
	return AM.ParentCtx
}
