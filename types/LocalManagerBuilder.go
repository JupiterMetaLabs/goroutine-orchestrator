package types

import (
	"sync"

	"github.com/neerajchowdary889/GoRoutinesManager/types/Errors"
)

const(
	Prefix_LocalManager = "LocalManager."
)

func NewLocalManager(localName string, appName string) *LocalManager {
	if IsIntilized().Local(localName, appName) {
		LocalManager, err := NewAppManager(appName).GetLocalManager(localName)
		if err != nil {
			return nil
		}
		return LocalManager
	}
	ctx, _ := NewAppManager(appName).GetAppContext()

	return &LocalManager{
		LocalName: localName,
		Routines: make(map[string]*Routine),
		ParentCtx: ctx,
	}

}

// Lock APIs
// LockAppReadMutex locks the app read mutex for the app manager - This is used to read the app manager's data
func (LM *LocalManager) LockAppReadMutex() {
	if LM.LocalMu == nil {
		LM.SetLocalMutex()
	}
	LM.LocalMu.RLock()
}

// UnlockAppReadMutex unlocks the app read mutex for the app manager - This is used to read the app manager's data
func (LM *LocalManager) UnlockAppReadMutex() {
	if LM.LocalMu == nil {
		LM.SetLocalMutex()
	}
	LM.LocalMu.RUnlock()
}

// LockAppWriteMutex locks the app write mutex for the app manager - This is used to write the app manager's data
func (LM *LocalManager) LockAppWriteMutex() {
	if LM.LocalMu == nil {
		LM.SetLocalMutex()
	}
	LM.LocalMu.Lock()
}

// UnlockAppWriteMutex unlocks the app write mutex for the app manager - This is used to write the app manager's data
func (LM *LocalManager) UnlockAppWriteMutex() {
	if LM.LocalMu == nil {
		LM.SetLocalMutex()
	}
	LM.LocalMu.Unlock()
}


func (LM *LocalManager) SetLocalMutex() *LocalManager {
	LM.LocalMu = &sync.RWMutex{}
	return LM
}


// Set APIs
// SetLocalName sets the name of the local manager
func (LM *LocalManager) SetLocalName(localName string) *LocalManager {
	LM.LocalName = localName
	return LM
}