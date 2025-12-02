package metrics

import (
	"runtime"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

// Collector is responsible for collecting metrics from the goroutine manager
type Collector struct {
	// stopCh is used to signal the collector to stop
	stopCh chan struct{}

	// intervalCh is used to signal interval changes
	intervalCh chan time.Duration

	// running indicates if the collector is currently running
	running bool

	// currentInterval stores the current interval for comparison
	currentInterval time.Duration
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		stopCh:          make(chan struct{}),
		intervalCh:      make(chan time.Duration, 1), // Buffered to avoid blocking
		running:         false,
		currentInterval: types.UpdateInterval,
	}
}

// Start begins collecting metrics at the configured interval
func (c *Collector) Start() {
	if c.running {
		return
	}

	c.running = true
	go c.collectLoop()
}

// Stop stops the metrics collector
func (c *Collector) Stop() {
	if !c.running {
		return
	}

	close(c.stopCh)
	c.running = false
}

// UpdateInterval updates the collection interval dynamically
// This implements the observer pattern for interval changes
func (c *Collector) UpdateInterval(newInterval time.Duration) {
	if !c.running {
		return
	}
	// Non-blocking send (channel is buffered)
	select {
	case c.intervalCh <- newInterval:
	default:
		// Channel full, skip (will be picked up on next tick check)
	}
}

// collectLoop runs the collection loop with dynamic interval support
// It observes changes to types.UpdateInterval and updates the ticker accordingly
func (c *Collector) collectLoop() {
    globalMgr, _ := types.GetGlobalManager()
    metadata := globalMgr.GetMetadata()
    
    currentInterval := metadata.GetUpdateInterval()
    c.currentInterval = currentInterval
    ticker := time.NewTicker(currentInterval)
    defer ticker.Stop()
    
    c.Collect()
    
    for {
        select {
        case <-ticker.C:
            c.Collect()
            
            newInterval := metadata.GetUpdateInterval()
            if newInterval != currentInterval {
                ticker.Stop()
                currentInterval = newInterval
                c.currentInterval = currentInterval
                ticker = time.NewTicker(currentInterval)
            }
        case newInterval := <-c.intervalCh:
            if newInterval != currentInterval {
                ticker.Stop()
                currentInterval = newInterval
                c.currentInterval = currentInterval
                ticker = time.NewTicker(currentInterval)
            }
        case <-c.stopCh:
            return
        }
    }
}

// Collect gathers all metrics from the goroutine manager
func (c *Collector) Collect() {
	if !IsInitialized() {
		return
	}

	c.collectGlobalMetrics()
	c.collectAppMetrics()
	c.collectLocalMetrics()
	c.collectGoroutineMetrics()
	c.collectMetadataMetrics()
	c.collectSystemMetrics()
}

// collectGlobalMetrics collects metrics from the global manager
func (c *Collector) collectGlobalMetrics() {
	// Check if global manager is initialized
	if types.IsIntilized().Global() {
		GlobalInitialized.Set(1)

		globalMgr, err := types.GetGlobalManager()
		if err != nil {
			GlobalInitialized.Set(0)
			return
		}

		// Count app managers
		appCount := globalMgr.GetAppManagerCount()
		AppManagersTotal.Set(float64(appCount))

		// Count local managers (efficient - doesn't create slices)
		localCount := 0
		appManagers := globalMgr.GetAppManagers()
		for _, appMgr := range appManagers {
			localCount += appMgr.GetLocalManagerCount()
		}
		LocalManagersTotal.Set(float64(localCount))

		// Count goroutines (efficient - doesn't create slices)
		goroutineCount := 0
		for _, appMgr := range appManagers {
			localManagers := appMgr.GetLocalManagers()
			for _, localMgr := range localManagers {
				goroutineCount += localMgr.GetRoutineCount()
			}
		}
		GoroutinesTotal.Set(float64(goroutineCount))

		// Get shutdown timeout
		metadata := globalMgr.GetMetadata()
		if metadata != nil {
			ShutdownTimeoutSeconds.Set(metadata.GetShutdownTimeout().Seconds())
		}
	} else {
		GlobalInitialized.Set(0)
		AppManagersTotal.Set(0)
		LocalManagersTotal.Set(0)
		GoroutinesTotal.Set(0)
	}
}

// collectAppMetrics collects metrics for each app manager
func (c *Collector) collectAppMetrics() {
	if !types.IsIntilized().Global() {
		return
	}

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		return
	}

	appManagers := globalMgr.GetAppManagers()

	// Track which apps we've seen to clean up old metrics
	seenApps := make(map[string]bool)

	for appName, appMgr := range appManagers {
		seenApps[appName] = true

		// App is initialized
		AppInitialized.WithLabelValues(appName).Set(1)

		// Count local managers
		localCount := appMgr.GetLocalManagerCount()
		AppLocalManagers.WithLabelValues(appName).Set(float64(localCount))

		// Count goroutines
		goroutineCount := 0
		localManagers := appMgr.GetLocalManagers()
		for _, localMgr := range localManagers {
			goroutineCount += localMgr.GetRoutineCount()
		}
		AppGoroutines.WithLabelValues(appName).Set(float64(goroutineCount))
	}
}

// collectLocalMetrics collects metrics for each local manager
func (c *Collector) collectLocalMetrics() {
	if !types.IsIntilized().Global() {
		return
	}

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		return
	}

	appManagers := globalMgr.GetAppManagers()

	for appName, appMgr := range appManagers {
		localManagers := appMgr.GetLocalManagers()

		for localName, localMgr := range localManagers {
			// Count goroutines
			goroutineCount := localMgr.GetRoutineCount()
			LocalGoroutines.WithLabelValues(appName, localName).Set(float64(goroutineCount))

			// Count function wait groups
			functionWgCount := localMgr.GetFunctionWgCount()
			LocalFunctionWaitgroups.WithLabelValues(appName, localName).Set(float64(functionWgCount))
		}
	}
}

// collectGoroutineMetrics collects detailed goroutine metrics
func (c *Collector) collectGoroutineMetrics() {
	if !types.IsIntilized().Global() {
		return
	}

	globalMgr, err := types.GetGlobalManager()
	if err != nil {
		return
	}

	appManagers := globalMgr.GetAppManagers()

	// Track goroutines by function
	functionCounts := make(map[string]map[string]map[string]int) // app -> local -> function -> count

	for appName, appMgr := range appManagers {
		localManagers := appMgr.GetLocalManagers()

		for localName, localMgr := range localManagers {
			routines := localMgr.GetRoutines()

			for _, routine := range routines {
				functionName := routine.FunctionName

				// Initialize maps if needed
				if functionCounts[appName] == nil {
					functionCounts[appName] = make(map[string]map[string]int)
				}
				if functionCounts[appName][localName] == nil {
					functionCounts[appName][localName] = make(map[string]int)
				}

				// Increment count
				functionCounts[appName][localName][functionName]++

				// Update goroutine age
				UpdateGoroutineAge(appName, localName, functionName, routine.ID, routine.StartedAt)
			}
		}
	}

	// Update function-based goroutine counts
	for appName, localMap := range functionCounts {
		for localName, functionMap := range localMap {
			for functionName, count := range functionMap {
				GoroutinesByFunction.WithLabelValues(appName, localName, functionName).Set(float64(count))
			}
		}
	}
}

// collectMetadataMetrics collects metrics from metadata
func (c *Collector) collectMetadataMetrics() {
    if !types.IsIntilized().Global() {
        MetricsEnabled.Set(0)
        MaxRoutines.Set(0)
        return
    }
    
    globalMgr, err := types.GetGlobalManager()
    if err != nil {
        return
    }
    
    metadata := globalMgr.GetMetadata()
    if metadata == nil {
        MetricsEnabled.Set(0)
        MaxRoutines.Set(0)
        return
    }
    
    // ✅ FIX: Use getter method
    if metadata.GetMetrics() {
        MetricsEnabled.Set(1)
    } else {
        MetricsEnabled.Set(0)
    }
    
    // ✅ FIX: Use getter method
    MaxRoutines.Set(float64(metadata.GetMaxRoutines()))
}

// collectSystemMetrics collects system-level metrics
func (c *Collector) collectSystemMetrics() {
	// Set build info (static, but we set it every time for consistency)
	goVersion := runtime.Version()
	BuildInfo.WithLabelValues("dev", goVersion).Set(1)
}
