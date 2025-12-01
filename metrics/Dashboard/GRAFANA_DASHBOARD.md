# Grafana Dashboard for GoRoutinesManager

This dashboard provides comprehensive visualization of all metrics exposed by the GoRoutinesManager library.

## Download the Dashboard

You can download the dashboard JSON file using wget:

```bash
wget https://raw.githubusercontent.com/neerajchowdary889/GoRoutinesManager/main/metrics/grafana-dashboard.json
```

Or using curl:

```bash
curl -O https://raw.githubusercontent.com/neerajchowdary889/GoRoutinesManager/main/metrics/grafana-dashboard.json
```

## Prerequisites

1. **Prometheus** - Must be configured to scrape your application's `/metrics` endpoint
2. **Grafana** - Version 8.0 or higher
3. **Prometheus Data Source** - Configured in Grafana pointing to your Prometheus instance

## Installation Steps

### 1. Download the Dashboard

```bash
wget https://raw.githubusercontent.com/neerajchowdary889/GoRoutinesManager/main/metrics/grafana-dashboard.json
```

### 2. Import into Grafana

1. Open Grafana in your browser
2. Navigate to **Dashboards** → **Import**
3. Click **Upload JSON file** and select the downloaded `grafana-dashboard.json`
4. Or paste the JSON content directly
5. Select your **Prometheus data source**
6. Click **Import**

### 3. Configure Prometheus Data Source

If you haven't set up Prometheus as a data source:

1. Go to **Configuration** → **Data Sources**
2. Click **Add data source**
3. Select **Prometheus**
4. Enter your Prometheus URL (e.g., `http://localhost:9090`)
5. Click **Save & Test**

## Dashboard Panels

The dashboard includes the following panels:

### Overview Panels (Top Row)
- **Global Manager Status** - Shows if the global manager is initialized
- **Total App Managers** - Count of app managers
- **Total Local Managers** - Count of local managers
- **Total Goroutines** - Total number of tracked goroutines

### Goroutine Metrics
- **Goroutines by App** - Time series of goroutines per app
- **Goroutines by Function** - Time series of goroutines grouped by function
- **Goroutine Duration Distribution** - Histogram of goroutine execution durations
- **Goroutine Age** - Age of currently running goroutines

### Operation Metrics
- **Goroutine Operations Rate** - Rate of goroutine operations (create, cancel, complete)
- **Manager Operations Rate** - Rate of manager operations (create, shutdown)
- **Operation Errors Rate** - Rate of operation errors
- **Operation Durations (p95/p99)** - Percentile latencies for goroutine operations
- **Manager Operation Durations (p95/p99)** - Percentile latencies for manager operations

### Shutdown Metrics
- **Shutdown Duration** - Histogram of shutdown operation durations
- **Goroutines Remaining After Shutdown** - Goroutines that didn't complete during shutdown

### Manager Metrics
- **Local Managers by App** - Count of local managers per app
- **Function Wait Groups** - Number of function wait groups per local manager

### Configuration
- **Metrics Enabled** - Whether metrics collection is enabled
- **Max Routines** - Configured maximum routines limit
- **Shutdown Timeout** - Configured shutdown timeout

## Prometheus Configuration

Ensure your `prometheus.yml` is configured to scrape your application:

```yaml
scrape_configs:
  - job_name: 'goroutines-manager'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']  # Your application's metrics endpoint
```

## Using the Dashboard

### Time Range
- Default: Last 15 minutes
- Adjustable via the time picker in the top right

### Refresh Rate
- Default: 10 seconds
- Adjustable via the refresh button

### Filtering
- Use Grafana's built-in query editor to filter by:
  - `app_name` - Filter by application name
  - `local_name` - Filter by local manager name
  - `function_name` - Filter by function name
  - `operation` - Filter by operation type

## Customization

You can customize the dashboard by:
1. Clicking on any panel title → **Edit**
2. Modifying the Prometheus queries
3. Adjusting visualization settings
4. Adding new panels for specific metrics

## Troubleshooting

### No Data Showing

1. **Check Prometheus is scraping**: Visit `http://your-prometheus:9090/targets`
2. **Verify metrics endpoint**: Visit `http://your-app:8080/metrics` and check for `goroutine_manager_*` metrics
3. **Check data source**: Verify Grafana can connect to Prometheus
4. **Verify time range**: Ensure the time range includes when metrics were collected

### Metrics Not Appearing

1. **Enable metrics**: Ensure `Metrics` is set to `true` in metadata
2. **Check collector**: Verify the metrics collector is running
3. **Check initialization**: Ensure `InitMetrics()` has been called

## Example Queries

You can use these PromQL queries directly in Grafana:

```promql
# Total goroutines
goroutine_manager_global_goroutines_total

# Goroutines by app
goroutine_manager_app_goroutines{app_name="myapp"}

# Operation rate
rate(goroutine_manager_operations_goroutine_operations_total[5m])

# Error rate
rate(goroutine_manager_operations_errors_total[5m])

# P95 latency
histogram_quantile(0.95, rate(goroutine_manager_operations_goroutine_operation_duration_seconds_bucket[5m]))
```

## Support

For issues or questions:
- Check the [metrics README](README.md)
- Review the [API documentation](API.md)
- Open an issue on GitHub

