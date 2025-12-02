package Managertests

import (
	"testing"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/App"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
	"github.com/neerajchowdary889/GoRoutinesManager/Tests/Common"
)

// resetGlobalState resets the global singleton for testing
func resetGlobalState() {
	Common.ResetGlobalState()
}

func TestGlobalManager_Init(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()

	// Test first initialization
	err := gm.Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify global manager was created
	if !types.IsIntilized().Global() {
		t.Error("Global manager should be initialized")
	}

	// Test idempotent initialization (should not error)
	err = gm.Init()
	if err != nil {
		t.Errorf("Second Init() should not error: %v", err)
	}
}

func TestGlobalManager_GetAllAppManagers_Empty(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	apps, err := gm.GetAllAppManagers()
	if err != nil {
		t.Fatalf("GetAllAppManagers() failed: %v", err)
	}

	if len(apps) != 0 {
		t.Errorf("Expected 0 app managers, got %d", len(apps))
	}
}

func TestGlobalManager_GetAppManagerCount(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Initially should be 0
	count := gm.GetAppManagerCount()
	if count != 0 {
		t.Errorf("Expected 0 app managers, got %d", count)
	}

	// Create an app manager
	appMgr := types.NewAppManager("test-app").SetAppContext().SetAppMutex()
	types.SetAppManager("test-app", appMgr)

	// Now should be 1
	count = gm.GetAppManagerCount()
	if count != 1 {
		t.Errorf("Expected 1 app manager, got %d", count)
	}
}

func TestGlobalManager_GetAllAppManagers_WithApps(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Create multiple app managers
	app1 := types.NewAppManager("app1").SetAppContext().SetAppMutex()
	app2 := types.NewAppManager("app2").SetAppContext().SetAppMutex()
	types.SetAppManager("app1", app1)
	types.SetAppManager("app2", app2)

	apps, err := gm.GetAllAppManagers()
	if err != nil {
		t.Fatalf("GetAllAppManagers() failed: %v", err)
	}

	if len(apps) != 2 {
		t.Errorf("Expected 2 app managers, got %d", len(apps))
	}
}

func TestGlobalManager_GetAllLocalManagers(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Create app manager
	appMgr := App.NewAppManager("test-app")
	app, _ := appMgr.CreateApp()

	// Create local managers using the Local manager interface
	localMgr1 := Local.NewLocalManager("test-app", "local1")
	localMgr2 := Local.NewLocalManager("test-app", "local2")

	local1, _ := localMgr1.CreateLocal("local1")
	local2, _ := localMgr2.CreateLocal("local2")

	app.AddLocalManager("local1", local1)
	app.AddLocalManager("local2", local2)

	// Get all local managers
	locals, err := gm.GetAllLocalManagers()
	if err != nil {
		t.Fatalf("GetAllLocalManagers() failed: %v", err)
	}

	if len(locals) != 2 {
		t.Errorf("Expected 2 local managers, got %d", len(locals))
	}
}

func TestGlobalManager_GetLocalManagerCount(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Initially should be 0
	count := gm.GetLocalManagerCount()
	if count != 0 {
		t.Errorf("Expected 0 local managers, got %d", count)
	}

	// Create app manager with local managers
	appMgr := App.NewAppManager("test-app")
	app, _ := appMgr.CreateApp()

	localMgr1 := Local.NewLocalManager("test-app", "local1")
	localMgr2 := Local.NewLocalManager("test-app", "local2")

	local1, _ := localMgr1.CreateLocal("local1")
	local2, _ := localMgr2.CreateLocal("local2")

	app.AddLocalManager("local1", local1)
	app.AddLocalManager("local2", local2)

	// Now should be 2
	count = gm.GetLocalManagerCount()
	if count != 2 {
		t.Errorf("Expected 2 local managers, got %d", count)
	}
}

func TestGlobalManager_GetAllGoroutines(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Create hierarchy: app -> local -> routines
	appMgr := App.NewAppManager("test-app")
	app, _ := appMgr.CreateApp()

	localMgr := Local.NewLocalManager("test-app", "local1")
	local, _ := localMgr.CreateLocal("local1")
	app.AddLocalManager("local1", local)

	// Create routines
	local.NewGoRoutine("func1")
	local.NewGoRoutine("func2")

	// Get all goroutines
	routines, err := gm.GetAllGoroutines()
	if err != nil {
		t.Fatalf("GetAllGoroutines() failed: %v", err)
	}

	if len(routines) != 2 {
		t.Errorf("Expected 2 routines, got %d", len(routines))
	}
}

func TestGlobalManager_GetGoroutineCount(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Initially should be 0
	count := gm.GetGoroutineCount()
	if count != 0 {
		t.Errorf("Expected 0 goroutines, got %d", count)
	}

	// Create hierarchy with routines
	appMgr := App.NewAppManager("test-app")
	app, _ := appMgr.CreateApp()

	localMgr := Local.NewLocalManager("test-app", "local1")
	local, _ := localMgr.CreateLocal("local1")
	app.AddLocalManager("local1", local)

	local.NewGoRoutine("func1")
	local.NewGoRoutine("func2")
	local.NewGoRoutine("func3")

	// Now should be 3
	count = gm.GetGoroutineCount()
	if count != 3 {
		t.Errorf("Expected 3 goroutines, got %d", count)
	}
}

func TestGlobalManager_MultipleAppsAndLocals(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Create multiple apps with multiple locals and routines
	// App1 with 2 locals
	app1Mgr := App.NewAppManager("app1")
	app1, _ := app1Mgr.CreateApp()

	local1_1Mgr := Local.NewLocalManager("app1", "local1")
	local1_2Mgr := Local.NewLocalManager("app1", "local2")
	local1_1, _ := local1_1Mgr.CreateLocal("local1")
	local1_2, _ := local1_2Mgr.CreateLocal("local2")

	app1.AddLocalManager("local1", local1_1)
	app1.AddLocalManager("local2", local1_2)

	// Add routines to app1 locals
	local1_1.NewGoRoutine("func1")
	local1_1.NewGoRoutine("func2")
	local1_2.NewGoRoutine("func3")

	// App2 with 1 local
	app2Mgr := App.NewAppManager("app2")
	app2, _ := app2Mgr.CreateApp()

	local2_1Mgr := Local.NewLocalManager("app2", "local1")
	local2_1, _ := local2_1Mgr.CreateLocal("local1")
	app2.AddLocalManager("local1", local2_1)

	// Add routines to app2 local
	local2_1.NewGoRoutine("func4")
	local2_1.NewGoRoutine("func5")

	// Verify counts
	appCount := gm.GetAppManagerCount()
	if appCount != 2 {
		t.Errorf("Expected 2 apps, got %d", appCount)
	}

	localCount := gm.GetLocalManagerCount()
	if localCount != 3 {
		t.Errorf("Expected 3 local managers, got %d", localCount)
	}

	routineCount := gm.GetGoroutineCount()
	if routineCount != 5 {
		t.Errorf("Expected 5 routines, got %d", routineCount)
	}
}

func TestGlobalManager_BeforeInit(t *testing.T) {
	resetGlobalState()

	gm := Global.NewGlobalManager()

	// Try to get app managers before init
	_, err := gm.GetAllAppManagers()
	if err == nil {
		t.Error("Expected error when getting app managers before init")
	}
}
