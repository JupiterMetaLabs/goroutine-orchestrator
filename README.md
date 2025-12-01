# GoRoutinesManager
A lightweight, hierarchical goroutine supervision system for Go. This package centralizes all goroutine creation, lifecycle control, and shutdown handling to prevent leaks, improve observability, and enforce safe concurrency patterns.

![GoRoutineManger](png/GRM.png "Architecture of the GoRoutinesManager")

## Features
- Centralized goroutine management
- Hierarchical structure for better organization
- Safe shutdown handling
- Prevents goroutine leaks
- Improved observability
- Enforces safe concurrency patterns
- Prometheus metrics integration

## Metrics Integration

This library exposes Prometheus metrics that can be scraped by Prometheus. **Prometheus uses a pull model** - it scrapes metrics from HTTP endpoints, not push.

### Quick Start

```go
import (
    "net/http"
    "time"
    
    "github.com/neerajchowdary889/GoRoutinesManager/metrics"
)

func main() {
    // Start the metrics collector
    metrics.StartCollector(5 * time.Second)
    
    // Register metrics handler with your HTTP server
    mux := http.NewServeMux()
    mux.Handle("/metrics", metrics.GetMetricsHandler())
    
    // Start your server
    http.ListenAndServe(":8080", mux)
}
```

Then configure Prometheus to scrape `http://your-app:8080/metrics`.

For detailed metrics documentation, see [metrics/README.md](metrics/README.md).

For a complete API reference of all metrics functions, see [metrics/API.md](metrics/API.md).

## Grafana Dashboard

A pre-built Grafana dashboard is available for visualizing all metrics:

**Download:**
```bash
wget https://raw.githubusercontent.com/neerajchowdary889/GoRoutinesManager/main/metrics/grafana-dashboard.json
```

**Import into Grafana:**
1. Go to Dashboards → Import
2. Upload the JSON file
3. Select your Prometheus data source
4. Click Import

For detailed setup instructions, see [metrics/GRAFANA_DASHBOARD.md](metrics/GRAFANA_DASHBOARD.md).
