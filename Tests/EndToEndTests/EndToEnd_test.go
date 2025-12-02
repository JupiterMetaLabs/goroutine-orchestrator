package EndToEndTests

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/App"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Local"
	"github.com/neerajchowdary889/GoRoutinesManager/metrics"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

// TestComplexSystemWithMetrics is a comprehensive end-to-end test that:
// 1. Initializes Global Manager with metrics enabled
// 2. Creates 10 App Managers
// 3. Each App Manager has 20 Local Managers (total: 200 local managers)
// 4. Waits for user input before spawning goroutines
// 5. Spawns goroutines that run for 30 seconds
// 6. Handles Ctrl+C for safe shutdown
// 7. Keeps metrics server running for Prometheus/Grafana integration
func TestComplexSystemWithMetrics(t *testing.T) {
	// ============================================================================
	// PHASE 1: INITIALIZE GLOBAL MANAGER WITH METRICS
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PHASE 1: Initializing Global Manager with Metrics")
	fmt.Println(strings.Repeat("=", 80))

	// Reset global state
	types.Global = nil

	// Initialize Global Manager
	globalMgr := Global.NewGlobalManager()
	if err := globalMgr.Init(); err != nil {
		t.Fatalf("Failed to initialize Global Manager: %v", err)
	}
	fmt.Println("✓ Global Manager initialized")

	// Enable metrics with 2 second update interval
	metricsPort := ":19090"
	_, err := globalMgr.UpdateMetadata(Global.SET_METRICS_URL, []interface{}{true, "", 2 * time.Second})
	if err != nil {
		t.Fatalf("Failed to enable metrics: %v", err)
	}
	fmt.Println("✓ Metrics enabled")

	// Start metrics collector
	metrics.StartCollector()
	fmt.Println("✓ Metrics collector started")

	// Create and start HTTP server for metrics endpoint
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.GetMetricsHandler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	metricsServer := &http.Server{
		Addr:    metricsPort,
		Handler: mux,
	}

	// Start metrics server in background
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("✓ Metrics server started on %s/metrics\n", metricsPort)

	// ============================================================================
	// PHASE 2: CREATE APP MANAGERS AND LOCAL MANAGERS
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PHASE 2: Creating App Managers and Local Managers")
	fmt.Println(strings.Repeat("=", 80))

	// Store local managers for later goroutine spawning
	localManagers := make(map[string]Interface.LocalGoroutineManagerInterface)
	totalApps := 10
	localManagersPerApp := 20

	// Create 10 App Managers
	for appNum := 1; appNum <= totalApps; appNum++ {
		appName := fmt.Sprintf("app%d", appNum)
		fmt.Printf("Creating %s...\n", appName)

		// Create App Manager
		appMgr := App.NewAppManager(appName)
		_, err := appMgr.CreateApp()
		if err != nil {
			t.Fatalf("Failed to create %s: %v", appName, err)
		}

		// Create 20 Local Managers per App
		for localNum := 1; localNum <= localManagersPerApp; localNum++ {
			localName := fmt.Sprintf("local%d", localNum)

			// Create Local Manager
			localMgr := Local.NewLocalManager(appName, localName)
			_, err = localMgr.CreateLocal(localName)
			if err != nil {
				t.Fatalf("Failed to create %s/%s: %v", appName, localName, err)
			}

			// Store local manager for later
			key := fmt.Sprintf("%s/%s", appName, localName)
			localManagers[key] = localMgr
		}

		fmt.Printf("  ✓ %s: %d local managers created\n", appName, localManagersPerApp)
	}

	totalLocalManagers := totalApps * localManagersPerApp
	fmt.Printf("\n✓ Manager setup complete: %d apps, %d local managers (0 goroutines)\n", totalApps, totalLocalManagers)

	// ============================================================================
	// PHASE 3: WAIT FOR USER INPUT BEFORE SPAWNING GOROUTINES
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PHASE 3: Ready to Spawn Goroutines")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("📊 Metrics endpoint: http://localhost%s/metrics\n", metricsPort)
	fmt.Printf("📊 Prometheus should scrape: %s\n", metricsPort)
	fmt.Printf("📊 System ready: %d apps, %d local managers\n", totalApps, totalLocalManagers)
	fmt.Println("\n💡 You can now:")
	fmt.Println("   1. Start Prometheus: cd Tests/EndToEndTests && docker-compose up -d")
	fmt.Println("   2. Connect Grafana to Prometheus (http://localhost:9090)")
	fmt.Println("   3. Import Grafana dashboard from metrics/Dashboard/grafana-dashboard.json")
	fmt.Println("   4. View live metrics")
	fmt.Println("\n⚠️  Press Enter to spawn goroutines (or Ctrl+C to exit)...")

	// Wait for user input
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	fmt.Println()

	// ============================================================================
	// PHASE 4: SPAWN GOROUTINES
	// ============================================================================
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("PHASE 4: Spawning Goroutines")
	fmt.Println(strings.Repeat("=", 80))

	// Track execution counts for shutdown verification
	executionCounts := make(map[string]map[string]*atomic.Int32)
	goroutinesPerLocal := 20 // 10 workers + 10 processors

	// Spawn goroutines for all local managers
	for key, localMgr := range localManagers {
		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			continue
		}
		appName := parts[0]
		localName := parts[1]

		// Initialize execution counter
		if executionCounts[appName] == nil {
			executionCounts[appName] = make(map[string]*atomic.Int32)
		}
		counter := &atomic.Int32{}
		executionCounts[appName][localName] = counter

		// Spawn 10 worker goroutines
		for i := 0; i < 10; i++ {
			localMgr.Go("worker", func(ctx context.Context) error {
				// Run for 30 seconds or until context is cancelled
				deadline := time.Now().Add(30 * time.Second)
				for {
					select {
					case <-ctx.Done():
						counter.Add(1)
						return ctx.Err()
					case <-time.After(100 * time.Millisecond):
						// Check if 30 seconds have passed
						if time.Now().After(deadline) {
							counter.Add(1)
							return nil // Normal completion after 30 seconds
						}
						// Simulate work
					}
				}
			}, Local.WithTimeout(30*time.Second), Local.AddToWaitGroup("worker"))

			// Spawn 10 processor goroutines
			localMgr.Go("processor", func(ctx context.Context) error {
				// Run for 30 seconds or until context is cancelled
				deadline := time.Now().Add(30 * time.Second)
				for {
					select {
					case <-ctx.Done():
						counter.Add(1)
						return ctx.Err()
					case <-time.After(100 * time.Millisecond):
						// Check if 30 seconds have passed
						if time.Now().After(deadline) {
							counter.Add(1)
							return nil // Normal completion after 30 seconds
						}
						// Simulate work
					}
				}
			}, Local.WithTimeout(30*time.Second), Local.AddToWaitGroup("processor"))
		}
	}

	totalGoroutines := totalLocalManagers * goroutinesPerLocal
	fmt.Printf("✓ Spawned %d goroutines (%d per local manager)\n", totalGoroutines, goroutinesPerLocal)
	fmt.Printf("✓ Goroutines will run for 30 seconds or until shutdown\n")

	// ============================================================================
	// PHASE 5: SETUP SAFE SHUTDOWN HANDLER (CTRL+C)
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PHASE 5: System Running - Monitoring Metrics")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("📊 Metrics available at: http://localhost%s/metrics\n", metricsPort)
	fmt.Printf("📊 Total goroutines running: %d\n", totalGoroutines)
	fmt.Printf("⏱️  Goroutines will run for 30 seconds\n")
	fmt.Println("\n⚠️  Press Ctrl+C for safe shutdown")

	// Setup signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for either 30 seconds to pass or Ctrl+C
	go func() {
		<-sigChan
		fmt.Println("\n\n" + strings.Repeat("=", 80))
		fmt.Println("SHUTDOWN REQUESTED (Ctrl+C)")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("Initiating safe shutdown...")
	}()

	// Wait for 30 seconds or shutdown signal
	shutdownTimeout := 30 * time.Second
	shutdownTimer := time.NewTimer(shutdownTimeout)
	defer shutdownTimer.Stop()

	select {
	case <-shutdownTimer.C:
		fmt.Println("\n\n" + strings.Repeat("=", 80))
		fmt.Println("30 SECONDS ELAPSED - Initiating Safe Shutdown")
		fmt.Println(strings.Repeat("=", 80))
	case <-sigChan:
		// Shutdown signal received, handled in goroutine above
	}

	// ============================================================================
	// PHASE 6: SAFE SHUTDOWN
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PHASE 6: Safe Shutdown")
	fmt.Println(strings.Repeat("=", 80))

	// Shutdown Global Manager (this will shutdown all apps, locals, and goroutines)
	fmt.Println("Shutting down Global Manager...")

	err = globalMgr.Shutdown(true)
	if err != nil {
		fmt.Printf("⚠️  Shutdown warning: %v\n", err)
	} else {
		fmt.Println("✓ Global Manager shutdown complete")
	}

	// Give time for goroutines to finish and metrics to update
	// Keep collector running so metrics reflect the shutdown state
	fmt.Println("Waiting for metrics to update after shutdown...")
	time.Sleep(5 * time.Second) // Give collector time to update metrics to 0

	// Now stop metrics collector (after metrics have been updated)
	metrics.StopCollector()
	fmt.Println("✓ Metrics collector stopped")

	// Give one more second for final metrics update
	time.Sleep(1 * time.Second)

	// Shutdown metrics server (keep it running until the end so Grafana can see the final state)
	fmt.Println("Shutting down metrics server...")
	serverCtx, serverCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer serverCancel()

	if err := metricsServer.Shutdown(serverCtx); err != nil {
		fmt.Printf("⚠️  Metrics server shutdown warning: %v\n", err)
	} else {
		fmt.Println("✓ Metrics server shutdown complete")
	}

	// ============================================================================
	// PHASE 7: SUMMARY
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SHUTDOWN COMPLETE")
	fmt.Println(strings.Repeat("=", 80))

	// Count shutdown goroutines
	totalShutdown := int32(0)
	for appName, locals := range executionCounts {
		appTotal := int32(0)
		for localName, counter := range locals {
			count := counter.Load()
			appTotal += count
			totalShutdown += count
			if count > 0 {
				fmt.Printf("  %s/%s: %d goroutines shutdown\n", appName, localName, count)
			}
		}
		if appTotal > 0 {
			fmt.Printf("  %s: %d total goroutines shutdown\n", appName, appTotal)
		}
	}

	fmt.Printf("\n✓ Total goroutines shutdown: %d/%d\n", totalShutdown, totalGoroutines)
	fmt.Println("✅ Test completed successfully")
}
