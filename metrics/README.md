# Metrics Integration Guide

## Prometheus Architecture: Pull Model

Prometheus uses a **pull model**, not a push model. This means:
- Prometheus **scrapes** metrics from HTTP endpoints
- Applications **expose** metrics via HTTP endpoints
- Prometheus does **not** accept direct pushes from applications

## For Libraries: Best Practices

As a library, you should **NOT**:
- ❌ Start your own HTTP server (conflicts with application's server)
- ❌ Push metrics directly to Prometheus
- ❌ Assume you control the HTTP server lifecycle

Instead, you should:
- ✅ Expose an HTTP handler that applications can register
- ✅ Let the consuming application control the HTTP server
- ✅ Use Prometheus's default registry (which you're already doing)

## Usage Patterns

### Pattern 1: Integrate with Application's HTTP Server (Recommended)

```go
package main

import (
    "net/http"
    
    "github.com/yourusername/GoRoutinesManager/metrics"
    "github.com/yourusername/GoRoutinesManager/types"
)

func main() {
    // Initialize your goroutine manager
    builder := types.NewBuilderGlobalManager()
    // ... configure and build ...
    
    // Start the metrics collector (collects data periodically)
    metrics.StartCollector(5 * time.Second)
    
    // Get the metrics handler and register with your HTTP server
    mux := http.NewServeMux()
    mux.Handle("/metrics", metrics.GetMetricsHandler())
    
    // Register your other routes
    mux.HandleFunc("/api/users", handleUsers)
    
    // Start your application's HTTP server
    http.ListenAndServe(":8080", mux)
}
```

### Pattern 2: Standalone Metrics Server (Optional, Not Recommended for Libraries)

If you absolutely need a standalone server (e.g., for testing or demos), you can use `StartMetricsServer()`, but this is **not recommended** for production libraries:

```go
// This starts its own HTTP server - use with caution
err := metrics.StartMetricsServer(":9090", 5*time.Second)
if err != nil {
    log.Fatal(err)
}
```

**Why not recommended?**
- Conflicts with application's HTTP server
- Port management issues
- Lifecycle conflicts
- Not idiomatic for Go libraries

## Prometheus Configuration

In your `prometheus.yml`, configure scraping:

```yaml
scrape_configs:
  - job_name: 'my-app'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']  # Your application's server
```

Prometheus will scrape `http://localhost:8080/metrics` automatically.

## Pushgateway (For Short-Lived Jobs Only)

**Note:** Pushgateway exists for short-lived jobs (batch jobs, cron jobs) that can't be scraped. However:
- ❌ **Not recommended for libraries** - libraries should expose handlers
- ✅ Only use if you're building a CLI tool or batch job that exits quickly
- ✅ The consuming application would push to Pushgateway, not the library

If you need Pushgateway support, the consuming application would do:

```go
import "github.com/prometheus/client_golang/prometheus/push"

pusher := push.New("http://pushgateway:9091", "my-job")
pusher.Collector(metrics.GetRegistry())
pusher.Push()
```

But again, this is **not the library's responsibility** - it's the application's choice.

## Summary

1. **Libraries expose handlers** → Applications register them → Prometheus scrapes
2. Your `GetMetricsHandler()` is the correct approach ✅
3. `StartMetricsServer()` is optional but not recommended for library consumers
4. Prometheus pulls, it doesn't accept pushes (except via Pushgateway for special cases)

## API Reference

For a complete list of all available metrics APIs, see [API.md](API.md).

The API reference includes:
- All public functions with signatures and parameters
- Usage examples for each API
- Exported metrics variables
- Collector type methods
- Complete documentation of all 18 public APIs and 13 exported metrics

