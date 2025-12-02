# Metrics Package API Reference

This document lists all public APIs exposed by the `metrics` package.

## Table of Contents
1. [Initialization APIs](#initialization-apis)
2. [Server Management APIs](#server-management-apis)
3. [Collector Management APIs](#collector-management-apis)
4. [Metrics Recording APIs](#metrics-recording-apis)
5. [Registry APIs](#registry-apis)
6. [Status/Query APIs](#statusquery-apis)
7. [Exported Metrics Variables](#exported-metrics-variables)
8. [Collector Type](#collector-type)

---

## Initialization APIs

### `InitMetrics()`
Initializes and registers all Prometheus metrics. This function is safe to call multiple times (uses `sync.Once`).

**Signature:**
```go
func InitMetrics()
```

**Usage:**
```go
metrics.InitMetrics()
```

---

### `IsInitialized() bool`
Returns whether metrics have been initialized.

**Signature:**
```go
func IsInitialized() bool
```

**Returns:**
- `bool`: `true` if metrics are initialized, `false` otherwise

**Usage:**
```go
if metrics.IsInitialized() {
    // Metrics are ready
}
```

---

## Server Management APIs

### `StartMetricsServer(addr string, updateInterval time.Duration) error`
Starts an HTTP server to expose Prometheus metrics. This function starts its own HTTP server.

**Signature:**
```go
func StartMetricsServer(addr string, updateInterval time.Duration) error
```

**Parameters:**
- `addr`: Address to listen on (e.g., `":9090"` or `"localhost:9090"`)
- `updateInterval`: How often to collect metrics (0 = use default of 5 seconds)

**Returns:**
- `error`: Returns an error if the server is already running or if initialization fails

**Note:** For library usage, it's recommended to use `GetMetricsHandler()` instead and register it with your application's HTTP server.

**Usage:**
```go
err := metrics.StartMetricsServer(":9090", 5*time.Second)
if err != nil {
    log.Fatal(err)
}
```

---

### `StopMetricsServer(ctx context.Context) error`
Gracefully stops the metrics HTTP server.

**Signature:**
```go
func StopMetricsServer(ctx context.Context) error
```

**Parameters:**
- `ctx`: Context for graceful shutdown timeout

**Returns:**
- `error`: Returns an error if the server is not running or if shutdown fails

**Usage:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := metrics.StopMetricsServer(ctx)
if err != nil {
    log.Printf("Error stopping metrics server: %v", err)
}
```

---

### `GetMetricsHandler() http.Handler`
Returns the HTTP handler for Prometheus metrics. This allows users to integrate metrics into their own HTTP server.

**Signature:**
```go
func GetMetricsHandler() http.Handler
```

**Returns:**
- `http.Handler`: Prometheus metrics HTTP handler

**Usage:**
```go
mux := http.NewServeMux()
mux.Handle("/metrics", metrics.GetMetricsHandler())
http.ListenAndServe(":8080", mux)
```

---

## Collector Management APIs

### `StartCollector(updateInterval time.Duration)`
Starts the metrics collector without starting an HTTP server. This is useful when you want to integrate metrics into an existing HTTP server.

**Signature:**
```go
func StartCollector(updateInterval time.Duration)
```

**Parameters:**
- `updateInterval`: How often to collect metrics (0 = use default of 5 seconds)

**Usage:**
```go
metrics.StartCollector(5 * time.Second)
```

---

### `StopCollector()`
Stops the metrics collector.

**Signature:**
```go
func StopCollector()
```

**Usage:**
```go
metrics.StopCollector()
```

---

### `NewCollector(updateInterval time.Duration) *Collector`
Creates a new metrics collector instance.

**Signature:**
```go
func NewCollector(updateInterval time.Duration) *Collector
```

**Parameters:**
- `updateInterval`: How often to update metrics (0 = use default of 5 seconds)

**Returns:**
- `*Collector`: A new Collector instance

**Usage:**
```go
collector := metrics.NewCollector(5 * time.Second)
collector.Start()
```

---

## Metrics Recording APIs

### `RecordGoroutineCompletion(appName, localName, functionName string, startTime int64)`
Records the completion of a goroutine. This should be called when a goroutine finishes execution.

**Signature:**
```go
func RecordGoroutineCompletion(appName, localName, functionName string, startTime int64)
```

**Parameters:**
- `appName`: Name of the app manager
- `localName`: Name of the local manager
- `functionName`: Name of the function that ran in the goroutine
- `startTime`: Start time of the goroutine (nanoseconds since Unix epoch)

**Usage:**
```go
startTime := time.Now().UnixNano()
// ... goroutine execution ...
metrics.RecordGoroutineCompletion("myApp", "myLocal", "myFunction", startTime)
```

---

### `UpdateGoroutineAge(appName, localName, functionName, routineID string, startTime int64)`
Updates the age metric for a specific goroutine.

**Signature:**
```go
func UpdateGoroutineAge(appName, localName, functionName, routineID string, startTime int64)
```

**Parameters:**
- `appName`: Name of the app manager
- `localName`: Name of the local manager
- `functionName`: Name of the function
- `routineID`: Unique identifier for the routine
- `startTime`: Start time of the goroutine (nanoseconds since Unix epoch)

**Usage:**
```go
metrics.UpdateGoroutineAge("myApp", "myLocal", "myFunction", "routine-123", startTime)
```

---

### `RemoveGoroutineAge(appName, localName, functionName, routineID string)`
Removes the age metric for a specific goroutine.

**Signature:**
```go
func RemoveGoroutineAge(appName, localName, functionName, routineID string)
```

**Parameters:**
- `appName`: Name of the app manager
- `localName`: Name of the local manager
- `functionName`: Name of the function
- `routineID`: Unique identifier for the routine

**Usage:**
```go
metrics.RemoveGoroutineAge("myApp", "myLocal", "myFunction", "routine-123")
```

---

## Registry APIs

### `GetRegistry() *prometheus.Registry`
Returns the Prometheus registry. If metrics haven't been initialized, it returns the default registry.

**Signature:**
```go
func GetRegistry() *prometheus.Registry
```

**Returns:**
- `*prometheus.Registry`: The Prometheus metrics registry

**Usage:**
```go
registry := metrics.GetRegistry()
```

---

### `ResetMetrics()`
Resets all metrics to their initial state. This is primarily useful for testing.

**Signature:**
```go
func ResetMetrics()
```

**Usage:**
```go
metrics.ResetMetrics()
```

---

## Status/Query APIs

### `IsServerRunning() bool`
Returns whether the metrics server is currently running.

**Signature:**
```go
func IsServerRunning() bool
```

**Returns:**
- `bool`: `true` if the server is running, `false` otherwise

**Usage:**
```go
if metrics.IsServerRunning() {
    // Server is active
}
```

---

### `IsCollectorRunning() bool`
Returns whether the metrics collector is currently running.

**Signature:**
```go
func IsCollectorRunning() bool
```

**Returns:**
- `bool`: `true` if the collector is running, `false` otherwise

**Usage:**
```go
if metrics.IsCollectorRunning() {
    // Collector is active
}
```

---

## Exported Metrics Variables

These are Prometheus metric variables that can be accessed directly (though typically you'd use the recording APIs instead).

### Global Manager Metrics

- `GlobalInitialized` (`prometheus.Gauge`) - Whether the global manager is initialized (1 = yes, 0 = no)
- `AppManagersTotal` (`prometheus.Gauge`) - Total number of app managers
- `LocalManagersTotal` (`prometheus.Gauge`) - Total number of local managers across all apps
- `GoroutinesTotal` (`prometheus.Gauge`) - Total number of tracked goroutines
- `ShutdownTimeoutSeconds` (`prometheus.Gauge`) - Configured shutdown timeout in seconds

### App Manager Metrics (with labels)

- `AppLocalManagers` (`*prometheus.GaugeVec`) - Number of local managers per app
  - Labels: `app_name`
- `AppGoroutines` (`*prometheus.GaugeVec`) - Number of goroutines per app
  - Labels: `app_name`
- `AppInitialized` (`*prometheus.GaugeVec`) - Whether an app is initialized (1 = yes, 0 = no)
  - Labels: `app_name`

### Local Manager Metrics (with labels)

- `LocalGoroutines` (`*prometheus.GaugeVec`) - Number of goroutines per local manager
  - Labels: `app_name`, `local_name`
- `LocalFunctionWaitgroups` (`*prometheus.GaugeVec`) - Number of function wait groups per local manager
  - Labels: `app_name`, `local_name`

### Goroutine Metrics (with labels)

- `GoroutinesByFunction` (`*prometheus.GaugeVec`) - Number of goroutines grouped by function
  - Labels: `app_name`, `local_name`, `function_name`
- `GoroutineDuration` (`*prometheus.HistogramVec`) - Duration of goroutines from start to completion
  - Labels: `app_name`, `local_name`, `function_name`
  - Buckets: `.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300` seconds
- `GoroutineAge` (`*prometheus.GaugeVec`) - Age of currently running goroutines in seconds
  - Labels: `app_name`, `local_name`, `function_name`, `routine_id`

### Metadata Metrics

- `MaxRoutines` (`prometheus.Gauge`) - Configured maximum routines limit (0 = unlimited)
- `MetricsEnabled` (`prometheus.Gauge`) - Whether metrics collection is enabled (1 = yes, 0 = no)

### System Metrics

- `BuildInfo` (`*prometheus.GaugeVec`) - Build information for the goroutine manager
  - Labels: `version`, `go_version`

---

## Collector Type

### `Collector`
The `Collector` type is responsible for collecting metrics from the goroutine manager.

**Type Definition:**
```go
type Collector struct {
    // Private fields
}
```

### `Collector.Start()`
Begins collecting metrics at the configured interval.

**Signature:**
```go
func (c *Collector) Start()
```

**Usage:**
```go
collector := metrics.NewCollector(5 * time.Second)
collector.Start()
```

---

### `Collector.Stop()`
Stops the metrics collector.

**Signature:**
```go
func (c *Collector) Stop()
```

**Usage:**
```go
collector.Stop()
```

---

### `Collector.Collect()`
Gathers all metrics from the goroutine manager. This is called automatically by the collector's loop, but can be called manually for immediate collection.

**Signature:**
```go
func (c *Collector) Collect()
```

**Usage:**
```go
collector.Collect() // Manual collection
```

---

## Summary

**Total Public APIs: 18**

- **Initialization:** 2 APIs
- **Server Management:** 3 APIs
- **Collector Management:** 3 APIs
- **Metrics Recording:** 3 APIs
- **Registry:** 2 APIs
- **Status/Query:** 2 APIs
- **Collector Methods:** 3 APIs

**Exported Metrics Variables: 13**

- Global metrics: 5
- App metrics: 3
- Local metrics: 2
- Goroutine metrics: 3
- Metadata metrics: 2
- System metrics: 1

