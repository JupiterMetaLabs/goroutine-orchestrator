package Managertests

import (
	"testing"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/App"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

func TestAppManager_CreateApp(t *testing.T) {
	resetGlobalState()

	// Create app manager
	appMgr := App.NewAppManager("test-app")

	// Create the app (should auto-init global if needed)
	app, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	if app == nil {
		t.Fatal("CreateApp() returned nil app")
	}

	// Verify global was initialized
	if !types.IsIntilized().Global() {
		t.Error("Global manager should be initialized")
	}

	// Verify app was initialized
	if !types.IsIntilized().App("test-app") {
		t.Error("App manager should be initialized")
	}

	// Verify app name
	if app.GetAppName() != "test-app" {
		t.Errorf("Expected app name 'test-app', got '%s'", app.GetAppName())
	}
}

func TestAppManager_CreateApp_Idempotent(t *testing.T) {
	resetGlobalState()

	appMgr := App.NewAppManager("test-app")

	// Create app first time
	app1, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("First CreateApp() failed: %v", err)
	}

	// Create app second time (should return existing)
	app2, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("Second CreateApp() failed: %v", err)
	}

	// Should be the same instance (pointer equality)
	if app1 != app2 {
		t.Error("CreateApp() should return same instance on subsequent calls")
	}
}

func TestAppManager_CreateApp_AutoInitGlobal(t *testing.T) {
	resetGlobalState()

	// Create app manager without manually initializing global
	appMgr := App.NewAppManager("test-app")

	// Verify global is not initialized yet
	if types.IsIntilized().Global() {
		t.Error("Global should not be initialized yet")
	}

	// Create the app (should auto-init global)
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	// Verify global was auto-initialized
	if !types.IsIntilized().Global() {
		t.Error("Global manager should be auto-initialized by CreateApp")
	}
}

func TestAppManager_GetAllLocalManagers(t *testing.T) {
	resetGlobalState()

	appMgr := App.NewAppManager("test-app")
	app, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	// Initially should be empty
	locals, err := appMgr.GetAllLocalManagers()
	if err != nil {
		t.Fatalf("GetAllLocalManagers() failed: %v", err)
	}
	if len(locals) != 0 {
		t.Errorf("Expected 0 local managers, got %d", len(locals))
	}

	// Create local managers using the Local interface
	localMgr1 := Local.NewLocalManager("test-app", "local1")
	localMgr2 := Local.NewLocalManager("test-app", "local2")

	local1, _ := localMgr1.CreateLocal("local1")
	local2, _ := localMgr2.CreateLocal("local2")

	app.AddLocalManager("local1", local1)
	app.AddLocalManager("local2", local2)

	// Now should have 2
	locals, err = appMgr.GetAllLocalManagers()
	if err != nil {
		t.Fatalf("GetAllLocalManagers() failed: %v", err)
	}
	if len(locals) != 2 {
		t.Errorf("Expected 2 local managers, got %d", len(locals))
	}
}

func TestAppManager_GetLocalManagerCount(t *testing.T) {
	resetGlobalState()

	appMgr := App.NewAppManager("test-app")
	app, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	// Initially 0
	count := appMgr.GetLocalManagerCount()
	if count != 0 {
		t.Errorf("Expected 0 local managers, got %d", count)
	}

	// Add some using Local interface
	for i := 1; i <= 3; i++ {
		localName := "local" + string(rune('0'+i))
		localMgr := Local.NewLocalManager("test-app", localName)
		local, _ := localMgr.CreateLocal(localName)
		app.AddLocalManager(localName, local)
	}

	count = appMgr.GetLocalManagerCount()
	if count != 3 {
		t.Errorf("Expected 3 local managers, got %d", count)
	}
}

func TestAppManager_GetGoroutineCount(t *testing.T) {
	resetGlobalState()

	appMgr := App.NewAppManager("test-app")
	app, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	// Initially 0
	count := appMgr.GetGoroutineCount()
	if count != 0 {
		t.Errorf("Expected 0 goroutines, got %d", count)
	}

	// Create local manager and add routines
	localMgr := Local.NewLocalManager("test-app", "local1")
	local, _ := localMgr.CreateLocal("local1")
	app.AddLocalManager("local1", local)

	local.NewGoRoutine("func1")
	local.NewGoRoutine("func2")

	count = appMgr.GetGoroutineCount()
	if count != 2 {
		t.Errorf("Expected 2 goroutines, got %d", count)
	}
}

func TestAppManager_GetAllGoroutines(t *testing.T) {
	resetGlobalState()

	appMgr := App.NewAppManager("test-app")
	app, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	// Create local managers with routines
	localMgr1 := Local.NewLocalManager("test-app", "local1")
	localMgr2 := Local.NewLocalManager("test-app", "local2")

	local1, _ := localMgr1.CreateLocal("local1")
	local2, _ := localMgr2.CreateLocal("local2")

	app.AddLocalManager("local1", local1)
	app.AddLocalManager("local2", local2)

	// Add routines to local1
	local1.NewGoRoutine("func1")
	local1.NewGoRoutine("func2")

	// Add routines to local2
	local2.NewGoRoutine("func3")

	// Get all goroutines
	routines, err := appMgr.GetAllGoroutines()
	if err != nil {
		t.Fatalf("GetAllGoroutines() failed: %v", err)
	}

	if len(routines) != 3 {
		t.Errorf("Expected 3 goroutines, got %d", len(routines))
	}
}

func TestAppManager_BeforeCreateApp(t *testing.T) {
	resetGlobalState()

	appMgr := App.NewAppManager("test-app")

	// Try operations before creating app
	_, err := appMgr.GetAllLocalManagers()
	if err == nil {
		t.Error("Expected error when calling GetAllLocalManagers before CreateApp")
	}

	count := appMgr.GetLocalManagerCount()
	if count != 0 {
		t.Errorf("Expected 0 count before CreateApp, got %d", count)
	}
}

func TestAppManager_MultipleApps(t *testing.T) {
	resetGlobalState()

	// Initialize global
	gm := Global.NewGlobalManager()
	gm.Init()

	// Create multiple apps
	app1Mgr := App.NewAppManager("app1")
	app2Mgr := App.NewAppManager("app2")

	app1, err := app1Mgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp(app1) failed: %v", err)
	}

	app2, err := app2Mgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp(app2) failed: %v", err)
	}

	// Create locals in each app
	local1_1Mgr := Local.NewLocalManager("app1", "local1")
	local1_1, _ := local1_1Mgr.CreateLocal("local1")
	app1.AddLocalManager("local1", local1_1)

	local2_1Mgr := Local.NewLocalManager("app2", "local1")
	local2_2Mgr := Local.NewLocalManager("app2", "local2")
	local2_1, _ := local2_1Mgr.CreateLocal("local1")
	local2_2, _ := local2_2Mgr.CreateLocal("local2")
	app2.AddLocalManager("local1", local2_1)
	app2.AddLocalManager("local2", local2_2)

	// Verify counts are independent
	count1 := app1Mgr.GetLocalManagerCount()
	if count1 != 1 {
		t.Errorf("App1 should have 1 local manager, got %d", count1)
	}

	count2 := app2Mgr.GetLocalManagerCount()
	if count2 != 2 {
		t.Errorf("App2 should have 2 local managers, got %d", count2)
	}

	// Verify global sees all
	totalCount := gm.GetLocalManagerCount()
	if totalCount != 3 {
		t.Errorf("Global should see 3 total local managers, got %d", totalCount)
	}
}
