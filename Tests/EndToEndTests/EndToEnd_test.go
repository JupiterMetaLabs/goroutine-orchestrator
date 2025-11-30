package EndToEndTests

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/App"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

// resetGlobalState resets the global manager state between tests
func resetGlobalState() {
	types.Global = nil
}

// setupComplexSystem creates a full hierarchy:
// 1 GlobalManager -> 5 AppManagers -> 5 LocalManagers each -> 20 goroutines each
// Total: 5 apps * 5 locals * 20 routines = 500 goroutines
func setupComplexSystem() (map[string]map[string]*atomic.Int32, error) {
	// Track execution counts: app -> local -> counter
	executionCounts := make(map[string]map[string]*atomic.Int32)

	fmt.Println("\n=== Setting up complex system ===")

	// Create 5 app managers
	for appNum := 1; appNum <= 5; appNum++ {
		appName := fmt.Sprintf("app%d", appNum)
		fmt.Printf("Creating %s...\n", appName)

		appMgr := App.NewAppManager(appName)
		_, err := appMgr.CreateApp()
		if err != nil {
			return nil, fmt.Errorf("failed to create %s: %v", appName, err)
		}

		executionCounts[appName] = make(map[string]*atomic.Int32)

		// Create 5 local managers per app
		for localNum := 1; localNum <= 5; localNum++ {
			localName := fmt.Sprintf("local%d", localNum)

			localMgr := Local.NewLocalManager(appName, localName)
			_, err = localMgr.CreateLocal(localName)
			if err != nil {
				return nil, fmt.Errorf("failed to create %s/%s: %v", appName, localName, err)
			}

			counter := &atomic.Int32{}
			executionCounts[appName][localName] = counter

			// Spawn 20 goroutines per local manager
			// 10 with function "worker", 10 with function "processor"
			for i := 0; i < 10; i++ {
				// Worker goroutines
				localMgr.GoWithWaitGroup("worker", func(ctx context.Context) error {
					for {
						select {
						case <-ctx.Done():
							counter.Add(1)
							return ctx.Err()
						case <-time.After(10 * time.Millisecond):
							// Simulate work
						}
					}
				})

				// Processor goroutines
				localMgr.GoWithWaitGroup("processor", func(ctx context.Context) error {
					for {
						select {
						case <-ctx.Done():
							counter.Add(1)
							return ctx.Err()
						case <-time.After(10 * time.Millisecond):
							// Simulate work
						}
					}
				})
			}
		}
		fmt.Printf("  ✓ %s: 5 local managers, 100 goroutines\n", appName)
	}

	fmt.Println("✓ System setup complete: 5 apps, 25 local managers, 500 goroutines")
	// Get all the appmanagers using the globalmanager
	appManagers := types.Global.GetAppManagers()
	for _, appMgr := range appManagers {
		fmt.Println("AppManager: ", appMgr)
		fmt.Println("LocalManagers: ", appMgr.GetLocalManagers())
		fmt.Println("Goroutines: ", appMgr.GetLocalManagerCount())
	}
	return executionCounts, nil
}

func TestEndToEnd_FunctionWaitGroupShutdown(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST 1: Function Wait Group Shutdown")
	fmt.Println(strings.Repeat("=", 80))
	resetGlobalState()

	counts, err := setupComplexSystem()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Give goroutines time to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown only "worker" function in app1/local1
	fmt.Println("\n--- Shutting down 'worker' function in app1/local1 ---")
	localMgr := Local.NewLocalManager("app1", "local1")

	beforeCount := localMgr.GetFunctionGoroutineCount("worker")
	fmt.Printf("Before shutdown: %d worker goroutines\n", beforeCount)

	err = localMgr.ShutdownFunction("worker", 2*time.Second)
	if err != nil {
		t.Fatalf("ShutdownFunction failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	afterCount := localMgr.GetFunctionGoroutineCount("worker")
	processorCount := localMgr.GetFunctionGoroutineCount("processor")

	fmt.Printf("After shutdown: %d worker goroutines, %d processor goroutines\n", afterCount, processorCount)

	// Verify only workers in app1/local1 were shutdown
	shutdownCount := counts["app1"]["local1"].Load()
	if shutdownCount != 10 {
		t.Errorf("Expected 10 workers shutdown in app1/local1, got %d", shutdownCount)
	} else {
		fmt.Printf("✓ Exactly 10 workers shutdown in app1/local1\n")
	}

	// Verify processors still running
	if processorCount != 10 {
		t.Errorf("Expected 10 processors still running, got %d", processorCount)
	} else {
		fmt.Println("✓ All 10 processors still running")
	}

	// Verify other local managers unaffected
	for localNum := 2; localNum <= 5; localNum++ {
		localName := fmt.Sprintf("local%d", localNum)
		if counts["app1"][localName].Load() != 0 {
			t.Errorf("app1/%s should be unaffected but has %d shutdowns",
				localName, counts["app1"][localName].Load())
		}
	}
	fmt.Println("✓ Other local managers in app1 unaffected")

	// Verify other apps unaffected
	for appNum := 2; appNum <= 5; appNum++ {
		appName := fmt.Sprintf("app%d", appNum)
		for localNum := 1; localNum <= 5; localNum++ {
			localName := fmt.Sprintf("local%d", localNum)
			if counts[appName][localName].Load() != 0 {
				t.Errorf("%s/%s should be unaffected", appName, localName)
			}
		}
	}
	fmt.Println("✓ All other apps unaffected")

	fmt.Println("\n✅ TEST 1 PASSED: Function-level shutdown works correctly")
}

func TestEndToEnd_LocalManagerShutdown(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST 2: LocalManager Shutdown")
	fmt.Println(strings.Repeat("=", 80))
	resetGlobalState()

	counts, err := setupComplexSystem()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Shutdown app2/local3 completely
	fmt.Println("\n--- Shutting down app2/local3 ---")
	localMgr := Local.NewLocalManager("app2", "local3")

	err = localMgr.Shutdown(true)
	if err != nil {
		t.Fatalf("LocalManager shutdown failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify all 20 goroutines in app2/local3 were shutdown
	shutdownCount := counts["app2"]["local3"].Load()
	if shutdownCount != 20 {
		t.Errorf("Expected 20 goroutines shutdown in app2/local3, got %d", shutdownCount)
	} else {
		fmt.Printf("✓ All 20 goroutines in app2/local3 shutdown\n")
	}

	// Verify other local managers in app2 unaffected
	for localNum := 1; localNum <= 5; localNum++ {
		if localNum == 3 {
			continue
		}
		localName := fmt.Sprintf("local%d", localNum)
		if counts["app2"][localName].Load() != 0 {
			t.Errorf("app2/%s should be unaffected but has %d shutdowns",
				localName, counts["app2"][localName].Load())
		}
	}
	fmt.Println("✓ Other local managers in app2 unaffected")

	// Verify other apps unaffected
	for appNum := 1; appNum <= 5; appNum++ {
		if appNum == 2 {
			continue
		}
		appName := fmt.Sprintf("app%d", appNum)
		for localNum := 1; localNum <= 5; localNum++ {
			localName := fmt.Sprintf("local%d", localNum)
			if counts[appName][localName].Load() != 0 {
				t.Errorf("%s/%s should be unaffected", appName, localName)
			}
		}
	}
	fmt.Println("✓ All other apps unaffected")

	fmt.Println("\n✅ TEST 2 PASSED: LocalManager shutdown works correctly")
}

func TestEndToEnd_AppManagerShutdown(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST 3: AppManager Shutdown")
	fmt.Println(strings.Repeat("=", 80))
	resetGlobalState()

	counts, err := setupComplexSystem()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Shutdown app3 completely
	fmt.Println("\n--- Shutting down app3 (5 local managers, 100 goroutines) ---")
	appMgr := App.NewAppManager("app3")

	err = appMgr.Shutdown(true)
	if err != nil {
		t.Fatalf("AppManager shutdown failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Verify all 100 goroutines in app3 were shutdown
	totalShutdown := int32(0)
	for localNum := 1; localNum <= 5; localNum++ {
		localName := fmt.Sprintf("local%d", localNum)
		count := counts["app3"][localName].Load()
		totalShutdown += count
		fmt.Printf("  app3/%s: %d goroutines shutdown\n", localName, count)
	}

	if totalShutdown != 100 {
		t.Errorf("Expected 100 goroutines shutdown in app3, got %d", totalShutdown)
	} else {
		fmt.Printf("✓ All 100 goroutines in app3 shutdown\n")
	}

	// Verify other apps unaffected
	for appNum := 1; appNum <= 5; appNum++ {
		if appNum == 3 {
			continue
		}
		appName := fmt.Sprintf("app%d", appNum)
		for localNum := 1; localNum <= 5; localNum++ {
			localName := fmt.Sprintf("local%d", localNum)
			if counts[appName][localName].Load() != 0 {
				t.Errorf("%s/%s should be unaffected but has %d shutdowns",
					appName, localName, counts[appName][localName].Load())
			}
		}
	}
	fmt.Println("✓ All other apps (app1, app2, app4, app5) unaffected")

	fmt.Println("\n✅ TEST 3 PASSED: AppManager shutdown works correctly")
}

func TestEndToEnd_GlobalManagerShutdown(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST 4: GlobalManager Shutdown")
	fmt.Println(strings.Repeat("=", 80))
	resetGlobalState()

	counts, err := setupComplexSystem()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Shutdown everything
	fmt.Println("\n--- Shutting down GlobalManager (5 apps, 25 locals, 500 goroutines) ---")
	globalMgr := Global.NewGlobalManager()

	startTime := time.Now()
	err = globalMgr.Shutdown(true)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("GlobalManager shutdown failed: %v", err)
	}

	fmt.Printf("Shutdown completed in %v\n", elapsed)

	time.Sleep(200 * time.Millisecond)

	// Verify ALL 500 goroutines were shutdown
	totalShutdown := int32(0)
	for appNum := 1; appNum <= 5; appNum++ {
		appName := fmt.Sprintf("app%d", appNum)
		appTotal := int32(0)
		for localNum := 1; localNum <= 5; localNum++ {
			localName := fmt.Sprintf("local%d", localNum)
			count := counts[appName][localName].Load()
			appTotal += count
			totalShutdown += count
		}
		fmt.Printf("  %s: %d goroutines shutdown\n", appName, appTotal)
	}

	if totalShutdown != 500 {
		t.Errorf("Expected 500 goroutines shutdown, got %d", totalShutdown)
	} else {
		fmt.Printf("✓ All 500 goroutines shutdown successfully\n")
	}

	fmt.Println("\n✅ TEST 4 PASSED: GlobalManager shutdown works correctly")
}

func TestEndToEnd_FullScenario(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST 5: Full End-to-End Scenario")
	fmt.Println(strings.Repeat("=", 80))
	resetGlobalState()

	counts, err := setupComplexSystem()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	fmt.Println("\n--- Step 1: Shutdown function 'worker' in app1/local1 ---")
	local1 := Local.NewLocalManager("app1", "local1")
	local1.ShutdownFunction("worker", 2*time.Second)
	time.Sleep(100 * time.Millisecond)

	step1Count := counts["app1"]["local1"].Load()
	fmt.Printf("✓ Step 1: %d goroutines shutdown\n", step1Count)

	fmt.Println("\n--- Step 2: Shutdown entire app2/local2 ---")
	local2 := Local.NewLocalManager("app2", "local2")
	local2.Shutdown(true)
	time.Sleep(100 * time.Millisecond)

	step2Count := counts["app2"]["local2"].Load()
	fmt.Printf("✓ Step 2: %d goroutines shutdown\n", step2Count)

	fmt.Println("\n--- Step 3: Shutdown entire app4 ---")
	app4 := App.NewAppManager("app4")
	app4.Shutdown(true)
	time.Sleep(200 * time.Millisecond)

	step3Count := int32(0)
	for localNum := 1; localNum <= 5; localNum++ {
		localName := fmt.Sprintf("local%d", localNum)
		step3Count += counts["app4"][localName].Load()
	}
	fmt.Printf("✓ Step 3: %d goroutines shutdown\n", step3Count)

	fmt.Println("\n--- Step 4: Shutdown GlobalManager (remaining apps) ---")
	globalMgr := Global.NewGlobalManager()
	globalMgr.Shutdown(true)
	time.Sleep(200 * time.Millisecond)

	// Count total shutdowns
	totalShutdown := int32(0)
	for appNum := 1; appNum <= 5; appNum++ {
		appName := fmt.Sprintf("app%d", appNum)
		for localNum := 1; localNum <= 5; localNum++ {
			localName := fmt.Sprintf("local%d", localNum)
			totalShutdown += counts[appName][localName].Load()
		}
	}

	fmt.Printf("\n✓ Total goroutines shutdown across all steps: %d/500\n", totalShutdown)

	// Verify expected counts
	if step1Count != 10 {
		t.Errorf("Step 1: expected 10, got %d", step1Count)
	}
	if step2Count != 20 {
		t.Errorf("Step 2: expected 20, got %d", step2Count)
	}
	if step3Count != 100 {
		t.Errorf("Step 3: expected 100, got %d", step3Count)
	}
	if totalShutdown != 500 {
		t.Errorf("Total: expected 500, got %d", totalShutdown)
	}

	fmt.Println("\n✅ TEST 5 PASSED: Full end-to-end scenario works correctly")
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("ALL END-TO-END TESTS PASSED! 🎉")
	fmt.Println(strings.Repeat("=", 80))
}

// Test the appmanager situation
func SimulateScenario(t *testing.T) {
	resetGlobalState()

	setupComplexSystem()

	fmt.Println("----< Scenario Setup Complete >----")
}
