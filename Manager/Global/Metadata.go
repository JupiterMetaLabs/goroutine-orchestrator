package Global

import (
	"errors"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Context"
	"github.com/neerajchowdary889/GoRoutinesManager/metrics"
	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

const (
	SET_METRICS_URL      = "SET_METRICS_URL"
	SET_SHUTDOWN_TIMEOUT = "SET_SHUTDOWN_TIMEOUT"
	SET_MAX_ROUTINES     = "SET_MAX_ROUTINES"
	SET_UPDATE_INTERVAL  = "SET_UPDATE_INTERVAL"
)

type metricsConfig struct {
	Enabled  bool
	URL      string
	Interval time.Duration
}

func (GM *GlobalManager) UpdateGlobalMetadata(flag string, value interface{}) (*types.Metadata, error) {
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
				ctx, cancel := Context.SpawnChild(g.Ctx)
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

func (GM *GlobalManager) GetGlobalMetadata() (*types.Metadata, error) {
	g, err := types.GetGlobalManager()
	if err != nil {
		return nil, err
	}
	return g.GetMetadata(), nil
}
