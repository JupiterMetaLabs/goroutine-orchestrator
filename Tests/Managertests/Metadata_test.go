package Managertests

import (
	"testing"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Global"
	"github.com/neerajchowdary889/GoRoutinesManager/Tests/Common"
	"github.com/neerajchowdary889/GoRoutinesManager/metrics"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

// TestMetadata_NewMetadata tests the creation of new metadata
func TestMetadata_NewMetadata(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		t.Fatalf("Failed to get global manager: %v", err)
	}

	// Create new metadata
	metadata := globalMgr.NewMetadata()
	if metadata == nil {
		t.Fatal("NewMetadata() returned nil")
	}

	// Verify default values
	if metadata.Metrics != false {
		t.Errorf("Expected Metrics to be false, got %v", metadata.Metrics)
	}

	if metadata.ShutdownTimeout != 10*time.Second {
		t.Errorf("Expected ShutdownTimeout to be 10s, got %v", metadata.ShutdownTimeout)
	}

	if metadata.MaxRoutines != 0 {
		t.Errorf("Expected MaxRoutines to be 0, got %d", metadata.MaxRoutines)
	}

	// Verify metadata is set in global manager
	retrievedMetadata := globalMgr.GetMetadata()
	if retrievedMetadata == nil {
		t.Error("Metadata not set in global manager")
	}
}

// TestMetadata_NewMetadata_Idempotent tests that NewMetadata returns existing metadata
func TestMetadata_NewMetadata_Idempotent(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		t.Fatalf("Failed to get global manager: %v", err)
	}

	// Create metadata first time
	metadata1 := globalMgr.NewMetadata()

	// Create metadata second time - should return same instance
	metadata2 := globalMgr.NewMetadata()

	if metadata1 != metadata2 {
		t.Error("NewMetadata() should return the same instance when called multiple times")
	}
}

// TestMetadata_SetMaxRoutines tests setting max routines
func TestMetadata_SetMaxRoutines(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		t.Fatalf("Failed to get global manager: %v", err)
	}

	metadata := globalMgr.NewMetadata()

	// Set max routines
	result := metadata.SetMaxRoutines(100)

	// Verify it returns the metadata for chaining
	if result != metadata {
		t.Error("SetMaxRoutines should return the same metadata instance for chaining")
	}

	// Verify value was set
	if metadata.MaxRoutines != 100 {
		t.Errorf("Expected MaxRoutines to be 100, got %d", metadata.MaxRoutines)
	}
}

// TestMetadata_SetShutdownTimeout tests setting shutdown timeout
func TestMetadata_SetShutdownTimeout(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		t.Fatalf("Failed to get global manager: %v", err)
	}

	metadata := globalMgr.NewMetadata()

	// Set shutdown timeout
	customTimeout := 30 * time.Second
	result := metadata.SetShutdownTimeout(customTimeout)

	// Verify it returns the metadata for chaining
	if result != metadata {
		t.Error("SetShutdownTimeout should return the same metadata instance for chaining")
	}

	// Verify value was set
	if metadata.ShutdownTimeout != customTimeout {
		t.Errorf("Expected ShutdownTimeout to be %v, got %v", customTimeout, metadata.ShutdownTimeout)
	}

	// Verify global variable was updated
	if types.ShutdownTimeout != customTimeout {
		t.Errorf("Expected global ShutdownTimeout to be %v, got %v", customTimeout, types.ShutdownTimeout)
	}
}

// TestMetadata_SetMetrics tests setting metrics configuration
func TestMetadata_SetMetrics(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		t.Fatalf("Failed to get global manager: %v", err)
	}

	metadata := globalMgr.NewMetadata()

	// Set metrics
	testURL := "http://localhost:9090/metrics"
	result := metadata.SetMetrics(true, testURL, 5*time.Second)

	// Verify it returns the metadata for chaining
	if result != metadata {
		t.Error("SetMetrics should return the same metadata instance for chaining")
	}

	// Verify values were set
	if !metadata.Metrics {
		t.Error("Expected Metrics to be true")
	}

	if metadata.MetricsURL != testURL {
		t.Errorf("Expected MetricsURL to be %s, got %s", testURL, metadata.MetricsURL)
	}
}

// TestMetadata_GetMetadata tests getting metadata
func TestMetadata_GetMetadata(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		t.Fatalf("Failed to get global manager: %v", err)
	}

	metadata := globalMgr.NewMetadata()

	// Get metadata
	retrieved := metadata.GetMetadata()

	if retrieved != metadata {
		t.Error("GetMetadata should return the same metadata instance")
	}
}

// TestMetadata_MethodChaining tests that all setter methods can be chained
func TestMetadata_MethodChaining(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		t.Fatalf("Failed to get global manager: %v", err)
	}

	metadata := globalMgr.NewMetadata()

	// Chain all setters
	result := metadata.
		SetMaxRoutines(200).
		SetShutdownTimeout(20*time.Second).
		SetMetrics(true, "http://metrics.example.com", 5*time.Second)

	// Verify all values were set
	if metadata.MaxRoutines != 200 {
		t.Errorf("Expected MaxRoutines to be 200, got %d", metadata.MaxRoutines)
	}

	if metadata.ShutdownTimeout != 20*time.Second {
		t.Errorf("Expected ShutdownTimeout to be 20s, got %v", metadata.ShutdownTimeout)
	}

	if !metadata.Metrics {
		t.Error("Expected Metrics to be true")
	}

	if metadata.MetricsURL != "http://metrics.example.com" {
		t.Errorf("Expected MetricsURL to be http://metrics.example.com, got %s", metadata.MetricsURL)
	}

	if result != metadata {
		t.Error("Method chaining should return the same metadata instance")
	}
}

// TestGlobalManager_UpdateMetadata_SetMaxRoutines tests updating max routines via GlobalManager
func TestGlobalManager_UpdateMetadata_SetMaxRoutines(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Test with int
	metadata, err := gm.UpdateMetadata(Global.SET_MAX_ROUTINES, 150)
	if err != nil {
		t.Fatalf("UpdateMetadata failed: %v", err)
	}

	if metadata.MaxRoutines != 150 {
		t.Errorf("Expected MaxRoutines to be 150, got %d", metadata.MaxRoutines)
	}

	// Test with int32
	metadata, err = gm.UpdateMetadata(Global.SET_MAX_ROUTINES, int32(200))
	if err != nil {
		t.Fatalf("UpdateMetadata with int32 failed: %v", err)
	}

	if metadata.MaxRoutines != 200 {
		t.Errorf("Expected MaxRoutines to be 200, got %d", metadata.MaxRoutines)
	}

	// Test with int64
	metadata, err = gm.UpdateMetadata(Global.SET_MAX_ROUTINES, int64(250))
	if err != nil {
		t.Fatalf("UpdateMetadata with int64 failed: %v", err)
	}

	if metadata.MaxRoutines != 250 {
		t.Errorf("Expected MaxRoutines to be 250, got %d", metadata.MaxRoutines)
	}

	// Test with *int
	maxRoutines := 300
	metadata, err = gm.UpdateMetadata(Global.SET_MAX_ROUTINES, &maxRoutines)
	if err != nil {
		t.Fatalf("UpdateMetadata with *int failed: %v", err)
	}

	if metadata.MaxRoutines != 300 {
		t.Errorf("Expected MaxRoutines to be 300, got %d", metadata.MaxRoutines)
	}
}

// TestGlobalManager_UpdateMetadata_SetShutdownTimeout tests updating shutdown timeout via GlobalManager
func TestGlobalManager_UpdateMetadata_SetShutdownTimeout(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Test with time.Duration
	timeout := 25 * time.Second
	metadata, err := gm.UpdateMetadata(Global.SET_SHUTDOWN_TIMEOUT, timeout)
	if err != nil {
		t.Fatalf("UpdateMetadata failed: %v", err)
	}

	if metadata.ShutdownTimeout != timeout {
		t.Errorf("Expected ShutdownTimeout to be %v, got %v", timeout, metadata.ShutdownTimeout)
	}

	// Test with *time.Duration
	timeout2 := 35 * time.Second
	metadata, err = gm.UpdateMetadata(Global.SET_SHUTDOWN_TIMEOUT, &timeout2)
	if err != nil {
		t.Fatalf("UpdateMetadata with *time.Duration failed: %v", err)
	}

	if metadata.ShutdownTimeout != timeout2 {
		t.Errorf("Expected ShutdownTimeout to be %v, got %v", timeout2, metadata.ShutdownTimeout)
	}
}

// TestGlobalManager_UpdateMetadata_SetMetricsURL tests updating metrics via GlobalManager
func TestGlobalManager_UpdateMetadata_SetMetricsURL(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Test with string (should enable metrics)
	testURL := "http://localhost:9090/metrics"
	metadata, err := gm.UpdateMetadata(Global.SET_METRICS_URL, testURL)
	if err != nil {
		t.Fatalf("UpdateMetadata with string failed: %v", err)
	}

	if !metadata.Metrics {
		t.Error("Expected Metrics to be enabled when passing string URL")
	}

	if metadata.MetricsURL != testURL {
		t.Errorf("Expected MetricsURL to be %s, got %s", testURL, metadata.MetricsURL)
	}

	// Test with slice []interface{}{bool, string}
	metadata, err = gm.UpdateMetadata(Global.SET_METRICS_URL, []interface{}{false, "http://disabled.com"})
	if err != nil {
		t.Fatalf("UpdateMetadata with slice failed: %v", err)
	}

	if metadata.Metrics {
		t.Error("Expected Metrics to be disabled")
	}

	if metadata.MetricsURL != "http://disabled.com" {
		t.Errorf("Expected MetricsURL to be http://disabled.com, got %s", metadata.MetricsURL)
	}

	// Test with array [2]interface{}{bool, string}
	metadata, err = gm.UpdateMetadata(Global.SET_METRICS_URL, [2]interface{}{true, "http://array.com"})
	if err != nil {
		t.Fatalf("UpdateMetadata with array failed: %v", err)
	}

	if !metadata.Metrics {
		t.Error("Expected Metrics to be enabled")
	}

	if metadata.MetricsURL != "http://array.com" {
		t.Errorf("Expected MetricsURL to be http://array.com, got %s", metadata.MetricsURL)
	}
}

// TestGlobalManager_UpdateMetadata_InvalidFlag tests error handling for invalid flags
func TestGlobalManager_UpdateMetadata_InvalidFlag(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	_, err := gm.UpdateMetadata("INVALID_FLAG", "value")
	if err == nil {
		t.Error("Expected error for invalid flag")
	}

	if err.Error() != "unknown update flag" {
		t.Errorf("Expected 'unknown update flag' error, got: %v", err)
	}
}

// TestGlobalManager_UpdateMetadata_InvalidTypes tests error handling for invalid value types
func TestGlobalManager_UpdateMetadata_InvalidTypes(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Invalid type for max routines
	_, err := gm.UpdateMetadata(Global.SET_MAX_ROUTINES, "not a number")
	if err == nil {
		t.Error("Expected error for invalid max routines type")
	}

	// Invalid type for shutdown timeout
	_, err = gm.UpdateMetadata(Global.SET_SHUTDOWN_TIMEOUT, 123)
	if err == nil {
		t.Error("Expected error for invalid shutdown timeout type")
	}

	// Invalid type for metrics URL
	_, err = gm.UpdateMetadata(Global.SET_METRICS_URL, 12345)
	if err == nil {
		t.Error("Expected error for invalid metrics URL type")
	}

	// Invalid slice length for metrics
	_, err = gm.UpdateMetadata(Global.SET_METRICS_URL, []interface{}{true})
	if err == nil {
		t.Error("Expected error for invalid metrics slice length")
	}

	// Invalid slice types for metrics
	_, err = gm.UpdateMetadata(Global.SET_METRICS_URL, []interface{}{"not bool", "url"})
	if err == nil {
		t.Error("Expected error for invalid metrics slice types")
	}
}

// TestGlobalManager_GetMetadata tests getting metadata via GlobalManager
func TestGlobalManager_GetMetadata(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Create metadata first
	globalMgr, _ := types.GetGlobalManager()
	expectedMetadata := globalMgr.NewMetadata()

	// Get metadata via GlobalManager interface
	metadata, err := gm.GetMetadata()
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if metadata != expectedMetadata {
		t.Error("GetMetadata should return the same metadata instance")
	}
}

// TestGlobalManager_GetMetadata_BeforeInit tests error handling when getting metadata before init
func TestGlobalManager_GetMetadata_BeforeInit(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()

	_, err := gm.GetMetadata()
	if err == nil {
		t.Error("Expected error when getting metadata before init")
	}
}

// TestGlobalManager_UpdateMetadata_BeforeInit tests error handling when updating metadata before init
func TestGlobalManager_UpdateMetadata_BeforeInit(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()

	_, err := gm.UpdateMetadata(Global.SET_MAX_ROUTINES, 100)
	if err == nil {
		t.Error("Expected error when updating metadata before init")
	}
}

// TestMetadata_ComplexScenario tests a complex scenario with multiple updates
func TestMetadata_ComplexScenario(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Set max routines
	metadata, err := gm.UpdateMetadata(Global.SET_MAX_ROUTINES, 500)
	if err != nil {
		t.Fatalf("Failed to set max routines: %v", err)
	}

	// Set shutdown timeout
	metadata, err = gm.UpdateMetadata(Global.SET_SHUTDOWN_TIMEOUT, 45*time.Second)
	if err != nil {
		t.Fatalf("Failed to set shutdown timeout: %v", err)
	}

	// Enable metrics
	metadata, err = gm.UpdateMetadata(Global.SET_METRICS_URL, "http://prometheus:9090/metrics")
	if err != nil {
		t.Fatalf("Failed to set metrics: %v", err)
	}

	// Verify all settings
	retrievedMetadata, err := gm.GetMetadata()
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	if retrievedMetadata.MaxRoutines != 500 {
		t.Errorf("Expected MaxRoutines to be 500, got %d", retrievedMetadata.MaxRoutines)
	}

	if retrievedMetadata.ShutdownTimeout != 45*time.Second {
		t.Errorf("Expected ShutdownTimeout to be 45s, got %v", retrievedMetadata.ShutdownTimeout)
	}

	if !retrievedMetadata.Metrics {
		t.Error("Expected Metrics to be enabled")
	}

	if retrievedMetadata.MetricsURL != "http://prometheus:9090/metrics" {
		t.Errorf("Expected MetricsURL to be http://prometheus:9090/metrics, got %s", retrievedMetadata.MetricsURL)
	}

	// Verify it's the same instance
	if metadata != retrievedMetadata {
		t.Error("Metadata instances should be the same")
	}
}

// TestGlobalManager_MetricsServer_Integration tests that the metrics server starts and stops
func TestGlobalManager_MetricsServer_Integration(t *testing.T) {
	Common.ResetGlobalState()

	gm := Global.NewGlobalManager()
	gm.Init()

	// Ensure server is not running initially
	if metrics.IsServerRunning() {
		t.Error("Metrics server should not be running initially")
	}

	// Enable metrics with a URL
	// Use a port that is unlikely to be in use
	testURL := ":19091"
	_, err := gm.UpdateMetadata(Global.SET_METRICS_URL, testURL)
	if err != nil {
		t.Fatalf("Failed to enable metrics: %v", err)
	}

	// Allow some time for server to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is running
	if !metrics.IsServerRunning() {
		t.Error("Metrics server should be running after enabling")
	}

	// Disable metrics
	_, err = gm.UpdateMetadata(Global.SET_METRICS_URL, []interface{}{false, ""})
	if err != nil {
		t.Fatalf("Failed to disable metrics: %v", err)
	}

	// Allow some time for server to stop
	time.Sleep(100 * time.Millisecond)

	// Verify server is stopped
	if metrics.IsServerRunning() {
		t.Error("Metrics server should be stopped after disabling")
	}
}
