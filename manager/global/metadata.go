package global

import (
	"errors"
	"time"

	"github.com/JupiterMetaLabs/goroutine-orchestrator/ctxo"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/metrics"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/types"
)

// Metadata configuration flags for UpdateMetadata() function.
// These constants define the available configuration options for the global manager.
const (
	// SET_METRICS_URL configures Prometheus metrics collection and server.
	// Accepted value types:
	//   - string: URL for metrics server (enables metrics with default interval)
	//   - []interface{}{bool, string}: [enabled, url] with default interval
	//   - []interface{}{bool, string, time.Duration}: [enabled, url, interval]
	//   - metricsConfig: struct with Enabled, URL, Interval fields
	// Example: UpdateMetadata(SET_METRICS_URL, ":9090")
	SET_METRICS_URL = "SET_METRICS_URL"

	// SET_SHUTDOWN_TIMEOUT configures the graceful shutdown timeout duration.
	// Accepted value types:
	//   - time.Duration: timeout duration
	//   - *time.Duration: pointer to timeout duration
	// Example: UpdateMetadata(SET_SHUTDOWN_TIMEOUT, 30*time.Second)
	SET_SHUTDOWN_TIMEOUT = "SET_SHUTDOWN_TIMEOUT"

	// SET_MAX_ROUTINES configures the maximum number of concurrent goroutines allowed.
	// This is a soft limit used for monitoring and alerting.
	// Accepted value types:
	//   - int, int32, int64: maximum goroutine count
	//   - *int: pointer to maximum goroutine count
	// Example: UpdateMetadata(SET_MAX_ROUTINES, 1000)
	SET_MAX_ROUTINES = "SET_MAX_ROUTINES"

	// SET_UPDATE_INTERVAL configures the metrics collection update interval.
	// This controls how frequently the metrics collector updates metric values.
	// Accepted value types:
	//   - time.Duration: update interval
	//   - *time.Duration: pointer to update interval
	// Example: UpdateMetadata(SET_UPDATE_INTERVAL, 10*time.Second)
	SET_UPDATE_INTERVAL = "SET_UPDATE_INTERVAL"
)

// metricsConfig is a structured configuration type for metrics settings.
// This provides a type-safe alternative to using slices or individual parameters.
type metricsConfig struct {
	// Enabled determines whether metrics collection is active
	Enabled bool
	// URL is the address for the metrics HTTP server (e.g., ":9090")
	URL string
	// Interval is the metrics collection update interval (0 uses default)
	Interval time.Duration
}

// UpdateGlobalMetadata updates global configuration settings using the specified flag and value.
// This method provides runtime configuration of timeouts, metrics, goroutine limits, and update intervals.
//
// Parameters:
//   - flag: Configuration flag constant (SET_METRICS_URL, SET_SHUTDOWN_TIMEOUT, SET_MAX_ROUTINES, SET_UPDATE_INTERVAL)
//   - value: Configuration value (type depends on flag, see flag constants for details)
//
// Supported Flags and Value Types:
//
//	SET_METRICS_URL:
//	  - string: ":9090" (enables metrics with default interval)
//	  - []interface{}{true, ":9090"} (enabled, url)
//	  - []interface{}{true, ":9090", 5*time.Second} (enabled, url, interval)
//	  - metricsConfig{Enabled: true, URL: ":9090", Interval: 5*time.Second}
//
//	SET_SHUTDOWN_TIMEOUT:
//	  - time.Duration: 30*time.Second
//
//	SET_MAX_ROUTINES:
//	  - int: 1000 (soft limit for monitoring)
//
//	SET_UPDATE_INTERVAL:
//	  - time.Duration: 10*time.Second (metrics collection frequency)
//
// Metrics Behavior:
//   - When enabled: Initializes metrics, starts collector/server (idempotent)
//   - When disabled: Stops collector and server if running
//   - Interval changes notify the running collector via observer pattern
//   - Server start is idempotent (won't fail if already running)
//
// Returns:
//   - *types.Metadata: Updated metadata configuration
//   - error: Returns error if flag is unknown or value type is incorrect
//
// Examples:
//
//	// Enable metrics with default interval
//	metadata, err := globalMgr.UpdateGlobalMetadata(SET_METRICS_URL, ":9090")
//
//	// Enable metrics with custom interval
//	metadata, err := globalMgr.UpdateGlobalMetadata(SET_METRICS_URL,
//	    []interface{}{true, ":9090", 5*time.Second})
//
//	// Set shutdown timeout
//	metadata, err := globalMgr.UpdateGlobalMetadata(SET_SHUTDOWN_TIMEOUT, 30*time.Second)
//
//	// Set max routines limit
//	metadata, err := globalMgr.UpdateGlobalMetadata(SET_MAX_ROUTINES, 1000)
func (GM *GlobalManagerStruct) UpdateGlobalMetadata(flag string, value interface{}) (*types.Metadata, error) {
	// Get the global manager first
	g, err := types.GetGlobalManager()
	if err != nil {
		return nil, err
	}

	metadata := g.NewMetadata()

	switch flag {
	case SET_METRICS_URL:
		// Accept several input shapes:
		//  - string -> URL (enabled = true, default interval)
		//  - metricsConfig (with Enabled, URL, Interval)
		//  - []interface{}{bool, string} or [2]interface{}{bool, string} (default interval)
		//  - []interface{}{bool, string, time.Duration} or [3]interface{}{bool, string, time.Duration} (with interval)
		var enabled bool
		var url string
		var interval = types.UpdateInterval

		switch v := value.(type) {
		case string:
			enabled = true
			url = v
			interval = types.UpdateInterval
			metadata.SetMetrics(enabled, url, interval)
		case metricsConfig:
			enabled = v.Enabled
			url = v.URL
			if v.Interval > 0 {
				interval = v.Interval
			} else {
				interval = types.UpdateInterval
			}
			metadata.SetMetrics(v.Enabled, v.URL, interval)
		case *metricsConfig:
			enabled = v.Enabled
			url = v.URL
			if v.Interval > 0 {
				interval = v.Interval
			} else {
				interval = types.UpdateInterval
			}
			metadata.SetMetrics(v.Enabled, v.URL, interval)
		case []interface{}:
			if len(v) == 2 {
				// [enabled(bool), url(string)] - use default interval
				enabledVal, ok1 := v[0].(bool)
				urlVal, ok2 := v[1].(string)
				if !ok1 || !ok2 {
					return nil, errors.New("metrics: expected [bool, string] in slice")
				}
				enabled = enabledVal
				url = urlVal
				interval = types.UpdateInterval
				metadata.SetMetrics(enabledVal, urlVal, interval)
			} else if len(v) == 3 {
				// [enabled(bool), url(string), interval(time.Duration)] - with interval
				enabledVal, ok1 := v[0].(bool)
				urlVal, ok2 := v[1].(string)
				intervalVal, ok3 := v[2].(time.Duration)
				if !ok1 || !ok2 || !ok3 {
					return nil, errors.New("metrics: expected [bool, string, time.Duration] in slice")
				}
				enabled = enabledVal
				url = urlVal
				interval = intervalVal
				metadata.SetMetrics(enabledVal, urlVal, interval)
			} else {
				return nil, errors.New("metrics: expected slice of length 2 or 3: [enabled(bool), url(string)] or [enabled(bool), url(string), interval(time.Duration)]")
			}
		case [2]interface{}:
			// [enabled(bool), url(string)] - use default interval
			enabledVal, ok1 := v[0].(bool)
			urlVal, ok2 := v[1].(string)
			if !ok1 || !ok2 {
				return nil, errors.New("metrics: expected [bool, string] array")
			}
			enabled = enabledVal
			url = urlVal
			interval = types.UpdateInterval
			metadata.SetMetrics(enabledVal, urlVal, interval)
		case [3]interface{}:
			// [enabled(bool), url(string), interval(time.Duration)] - with interval
			enabledVal, ok1 := v[0].(bool)
			urlVal, ok2 := v[1].(string)
			intervalVal, ok3 := v[2].(time.Duration)
			if !ok1 || !ok2 || !ok3 {
				return nil, errors.New("metrics: expected [bool, string, time.Duration] array")
			}
			enabled = enabledVal
			url = urlVal
			interval = intervalVal
			metadata.SetMetrics(enabledVal, urlVal, interval)
		default:
			return nil, errors.New("metrics: unsupported value type; expected string, metricsConfig, [bool,string], or [bool,string,time.Duration]")
		}

		// Handle metrics enable/disable
		if enabled {
			// Always initialize metrics if enabled (InitMetrics is idempotent)
			metrics.InitMetrics()
			// The interval is now stored in metadata and will be used by the collector
			// via types.UpdateInterval (set by SetMetrics)
			// Notify collector about interval change (Observer pattern)
			metrics.UpdateMetricsUpdateInterval()
			if url != "" {
				// Start the metrics server if a URL is provided
				// If server is already running, ignore the error (idempotent behavior)
				if err := metrics.StartMetricsServer(url); err != nil {
					// If server is already running, continue (metrics are already active)
					// Only return error if it's a different error
					if err.Error() != "metrics server is already running" {
						return nil, err
					}
				}
			} else {
				// Start the collector (metrics will be collected periodically)
				// The collector uses types.UpdateInterval which is set by SetMetrics
				// StartCollector is idempotent - it checks if collector is already running
				metrics.StartCollector()
			}
		} else {
			// Disable metrics only if they are currently enabled
			// Check if collector is running before stopping
			if metrics.IsCollectorRunning() {
				metrics.StopCollector()
			}
			// Check if server is running before stopping
			if metrics.IsServerRunning() {
				// Get the parent context for this - dont create a new context
				// spawn a child context from global context
				ctx, cancel := ctxo.SpawnChild(g.Ctx)
				// StopMetricsServer may return error if not running, but we checked, so ignore errors
				_ = metrics.StopMetricsServer(ctx)
				cancel()
			}
			// If metrics are not running, continue (no-op)
		}

	case SET_SHUTDOWN_TIMEOUT:
		switch t := value.(type) {
		case time.Duration:
			metadata.SetShutdownTimeout(t)
		case *time.Duration:
			metadata.SetShutdownTimeout(*t)
		default:
			return nil, errors.New("shutdown timeout: expected time.Duration")
		}

	case SET_MAX_ROUTINES:
		switch n := value.(type) {
		case int:
			metadata.SetMaxRoutines(n)
		case int32:
			metadata.SetMaxRoutines(int(n))
		case int64:
			metadata.SetMaxRoutines(int(n))
		case *int:
			metadata.SetMaxRoutines(*n)
		default:
			return nil, errors.New("max routines: expected integer type")
		}

	case SET_UPDATE_INTERVAL:
		switch t := value.(type) {
		case time.Duration:
			metadata.UpdateIntervalTime(t)
		case *time.Duration:
			metadata.UpdateIntervalTime(*t)
		default:
			return nil, errors.New("update interval: expected time.Duration")
		}

	default:
		return nil, errors.New("unknown update flag")
	}

	return metadata, nil
}

// GetGlobalMetadata retrieves the current global metadata configuration.
// This returns the metadata instance containing all configuration settings including
// shutdown timeout, metrics enablement, max routines limit, and update intervals.
//
// Returns:
//   - *types.Metadata: Current metadata configuration
//   - error: Returns error if global manager is not initialized
//
// Example:
//
//	metadata, err := globalMgr.GetGlobalMetadata()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	}
//	if metadata.GetMetrics() {
//	    log.Printf("Metrics enabled at: %s", metadata.GetMetricsURL())
//	    log.Printf("Update interval: %v", metadata.GetUpdateInterval())
//	}
//	log.Printf("Shutdown timeout: %v", metadata.GetShutdownTimeout())
func (GM *GlobalManagerStruct) GetGlobalMetadata() (*types.Metadata, error) {
	g, err := types.GetGlobalManager()
	if err != nil {
		return nil, err
	}
	return g.GetMetadata(), nil
}
