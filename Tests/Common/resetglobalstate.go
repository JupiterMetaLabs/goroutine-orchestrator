package Common

import (
		"sync"
	"time"

	"github.com/JupiterMetaLabs/goroutine-orchestrator/Context"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/metrics"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/types"
)

var (
    once sync.Once
    lock sync.RWMutex
)


// Tests/Common/resetglobalstate.go
func ResetGlobalState() {
    // Stop the metrics server properly
    if metrics.IsServerRunning() {
        ctx, cancel := Context.GetAppContext("Test:ResetGlobalState").NewChildContext()
        defer cancel()
        metrics.StopMetricsServer(ctx)
    }
    
    // Wait longer for all goroutines to finish
    time.Sleep(300 * time.Millisecond)
    
    // Acquire write lock before resetting
    lock.Lock()
    once = sync.Once{}
    types.Global = nil
    lock.Unlock()
}