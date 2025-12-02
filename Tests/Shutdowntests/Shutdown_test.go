package Shutdowntests

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/App"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/Tests/Common"
)

func TestLocalManager_SafeShutdown(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_SafeShutdown ===")
	Common.ResetGlobalState()

	// Setup
	appMgr := App.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := Local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Spawn goroutines with wait group
	fmt.Println("Spawning 5 goroutines with wait group...")
	var counter atomic.Int32
	for i := 0; i < 5; i++ {
		localMgr.Go("worker", func(ctx context.Context) error {
			counter.Add(1)
			time.Sleep(100 * time.Millisecond)
			return nil
		}, Local.AddToWaitGroup("worker"))
	}
	fmt.Println("✓ Goroutines spawned")

	// Safe shutdown - should wait for all to complete
	fmt.Println("Initiating safe shutdown...")
	startTime := time.Now()
	err = localMgr.Shutdown(true)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
	fmt.Printf("✓ Safe shutdown completed in %v\n", elapsed)

	// Verify all goroutines completed
	if counter.Load() != 5 {
		t.Errorf("Expected 5 goroutines to complete, got %d", counter.Load())
	} else {
		fmt.Println("✓ All goroutines completed gracefully")
	}

	// Should have waited at least 100ms
	if elapsed < 100*time.Millisecond {
		t.Error("Safe shutdown should have waited for goroutines")
	}
}

func TestLocalManager_UnsafeShutdown(t *testing.T) {
	fmt.Println("\n=== TestLocalManager_UnsafeShutdown ===")
	Common.ResetGlobalState()

	// Setup
	appMgr := App.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := Local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	// Spawn long-running goroutines
	fmt.Println("Spawning 5 long-running goroutines...")
	var cancelled atomic.Int32
	for i := 0; i < 5; i++ {
		localMgr.Go("worker", func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				cancelled.Add(1)
				return ctx.Err()
			case <-time.After(10 * time.Second):
				return nil
			}
		})
	}
	fmt.Println("✓ Goroutines spawned")

	// Give them time to start
	time.Sleep(50 * time.Millisecond)

	// Unsafe shutdown - should cancel immediately
	fmt.Println("Initiating unsafe shutdown...")
	startTime := time.Now()
	err = localMgr.Shutdown(false)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
	fmt.Printf("✓ Unsafe shutdown completed in %v\n", elapsed)

	// Wait a bit for cancellations to propagate
	time.Sleep(100 * time.Millisecond)

	// Verify goroutines were cancelled
	if cancelled.Load() != 5 {
		t.Errorf("Expected 5 goroutines to be cancelled, got %d", cancelled.Load())
	} else {
		fmt.Println("✓ All goroutines cancelled immediately")
	}

	// Should have been fast (not waited 10 seconds)
	if elapsed > 1*time.Second {
		t.Error("Unsafe shutdown should be immediate")
	}
}

func TestAppManager_SafeShutdown(t *testing.T) {
	fmt.Println("\n=== TestAppManager_SafeShutdown ===")
	Common.ResetGlobalState()

	// Setup
	appMgr := App.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	// Create multiple local managers
	fmt.Println("Creating 3 local managers with goroutines...")
	for i := 1; i <= 3; i++ {
		localName := fmt.Sprintf("local%d", i)
		localMgr := Local.NewLocalManager("test-app", localName)
		_, err = localMgr.CreateLocal(localName)
		if err != nil {
			t.Fatalf("CreateLocal() failed: %v", err)
		}

		// Spawn goroutines in each local manager
		for j := 0; j < 2; j++ {
			localMgr.Go("worker", func(ctx context.Context) error {
				time.Sleep(100 * time.Millisecond)
				return nil
			}, Local.AddToWaitGroup("worker"))
		}
	}
	fmt.Println("✓ 3 local managers created with 2 goroutines each")

	// Safe shutdown app
	fmt.Println("Initiating safe app shutdown...")
	startTime := time.Now()
	err = appMgr.Shutdown(true)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
	fmt.Printf("✓ Safe app shutdown completed in %v\n", elapsed)

	// Should have waited for all goroutines
	if elapsed < 100*time.Millisecond {
		t.Error("Safe shutdown should have waited for all local managers")
	} else {
		fmt.Println("✓ All local managers completed gracefully")
	}
}

func TestGlobalManager_SafeShutdown(t *testing.T) {
	fmt.Println("\n=== TestGlobalManager_SafeShutdown ===")
	Common.ResetGlobalState()

	// Create multiple apps with local managers
	fmt.Println("Creating 2 apps with local managers...")
	for appNum := 1; appNum <= 2; appNum++ {
		appName := fmt.Sprintf("app%d", appNum)
		appMgr := App.NewAppManager(appName)
		_, err := appMgr.CreateApp()
		if err != nil {
			t.Fatalf("CreateApp() failed: %v", err)
		}

		// Create local managers in each app
		for localNum := 1; localNum <= 2; localNum++ {
			localName := fmt.Sprintf("local%d", localNum)
			localMgr := Local.NewLocalManager(appName, localName)
			_, err = localMgr.CreateLocal(localName)
			if err != nil {
				t.Fatalf("CreateLocal() failed: %v", err)
			}

			// Spawn goroutines
			localMgr.Go("worker", func(ctx context.Context) error {
				time.Sleep(100 * time.Millisecond)
				return nil
			}, Local.AddToWaitGroup("worker"))
		}
	}
	fmt.Println("✓ 2 apps created with 2 local managers each")

	// Give goroutines time to start
	time.Sleep(10 * time.Millisecond)

	// Safe shutdown global
	fmt.Println("Initiating safe global shutdown...")
	globalMgr := Global.NewGlobalManager()
	startTime := time.Now()
	err := globalMgr.Shutdown(true)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
	fmt.Printf("✓ Safe global shutdown completed in %v\n", elapsed)
	fmt.Println("✓ All apps and local managers shutdown initiated")
}

func TestGlobalManager_UnsafeShutdown(t *testing.T) {
	fmt.Println("\n=== TestGlobalManager_UnsafeShutdown ===")
	Common.ResetGlobalState()

	// Create app with long-running goroutines
	fmt.Println("Creating app with long-running goroutines...")
	appMgr := App.NewAppManager("test-app")
	_, err := appMgr.CreateApp()
	if err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	localMgr := Local.NewLocalManager("test-app", "test-local")
	_, err = localMgr.CreateLocal("test-local")
	if err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	var cancelled atomic.Int32
	for i := 0; i < 5; i++ {
		localMgr.Go("worker", func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				cancelled.Add(1)
				return ctx.Err()
			case <-time.After(10 * time.Second):
				return nil
			}
		})
	}
	fmt.Println("✓ Long-running goroutines spawned")

	time.Sleep(50 * time.Millisecond)

	// Unsafe shutdown global
	fmt.Println("Initiating unsafe global shutdown...")
	globalMgr := Global.NewGlobalManager()
	startTime := time.Now()
	err = globalMgr.Shutdown(false)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
	fmt.Printf("✓ Unsafe global shutdown completed in %v\n", elapsed)

	// Wait for cancellations
	time.Sleep(100 * time.Millisecond)

	// Verify goroutines were cancelled
	if cancelled.Load() != 5 {
		t.Errorf("Expected 5 goroutines to be cancelled, got %d", cancelled.Load())
	} else {
		fmt.Println("✓ All goroutines cancelled immediately")
	}

	// Should have been fast
	if elapsed > 1*time.Second {
		t.Error("Unsafe shutdown should be immediate")
	}
}
