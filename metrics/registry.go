package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// defaultRegistry is the default Prometheus registry
	defaultRegistry *prometheus.Registry
)

// GetRegistry returns the Prometheus registry
// If metrics haven't been initialized, it returns the default registry
func GetRegistry() *prometheus.Registry {
	if defaultRegistry == nil {
		defaultRegistry = prometheus.DefaultRegisterer.(*prometheus.Registry)
	}
	return defaultRegistry
}

// ResetMetrics resets all metrics to their initial state
// This is primarily useful for testing
func ResetMetrics() {
	if !IsInitialized() {
		return
	}

	// Reset global metrics
	GlobalInitialized.Set(0)
	AppManagersTotal.Set(0)
	LocalManagersTotal.Set(0)
	GoroutinesTotal.Set(0)
	ShutdownTimeoutSeconds.Set(0)

	// Reset app metrics (delete all label combinations)
	AppLocalManagers.Reset()
	AppGoroutines.Reset()
	AppInitialized.Reset()

	// Reset local metrics
	LocalGoroutines.Reset()
	LocalFunctionWaitgroups.Reset()

	// Reset goroutine metrics
	GoroutinesByFunction.Reset()
	GoroutineDuration.Reset()
	GoroutineAge.Reset()

	// Reset metadata metrics
	MaxRoutines.Set(0)
	MetricsEnabled.Set(0)

	// Reset system metrics
	BuildInfo.Reset()
}
