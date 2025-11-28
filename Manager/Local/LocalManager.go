package Local

import (
	"context"
	"sync"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
	"github.com/neerajchowdary889/GoRoutinesManager/types/Errors"
)

type LocalManager struct{
	AppName string
	LocalName string
}

func NewLocalManager(appName, localName string) Interface.LocalGoroutineManagerInterface {
	return &LocalManager{
		AppName: appName,
		LocalName: localName,
	}
}

func (LM *LocalManager) CreateLocal(localName string) (*types.LocalManager, error) {
	// First get the app manager
	appManager, err := types.GetAppManager(LM.AppName)
	if err != nil {
		return nil, err
	}
	if !types.IsIntilized().Local(appManager.GetAppName(), LM.LocalName){
		localManager := types.NewLocalManager(LM.LocalName, appManager.GetAppName()).SetLocalContext().SetLocalMutex().SetLocalWaitGroup()
		if localManager == nil {
			return nil, Errors.ErrLocalManagerNotFound
		}
		appManager.AddLocalManager(LM.LocalName, localManager)
		return localManager, nil
	}
	return appManager.GetLocalManager(LM.LocalName)
}

// Shutdowner
func (LM *LocalManager) Shutdown(safe bool) error {
	//TODO
	return nil
}

// FunctionShutdowner
func (LM *LocalManager) ShutdownFunction(functionName string, timeout time.Duration) error {
	//TODO
	return nil
}

// GoroutineSpawner
func (LM *LocalManager) Go(functionName string, workerFunc func(ctx context.Context) error) error {
	//TODO
	return nil
}

// GoroutineLister
func (LM *LocalManager) GetAllGoroutines() ([]*types.Routine, error) {
	//TODO
	return nil, nil
}

func (LM *LocalManager) GetGoroutineCount() int {
	//TODO
	return 0
}

// FunctionWaitGroupCreator
func (LM *LocalManager) NewFunctionWaitGroup(ctx context.Context, functionName string) (*sync.WaitGroup, error) {
	//TODO
	return nil, nil
}