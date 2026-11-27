// Package nowaitgroupstuck holds the one test that leaves a goroutine stuck.
// WaitForFunctionWithTimeout keeps its internal waiter running after a
// timeout (documented in manager/local/routine.go), so this test is alone in
// its package: no later test resets the global state that waiter has read.
package nowaitgroupstuck_test

import (
	"context"
	"testing"
	"time"

	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/app"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/local"
	common "github.com/JupiterMetaLabs/goroutine-orchestrator/test/common"
)

// A goroutine without a wait group that ignores cancellation still makes
// WaitForFunctionWithTimeout time out: the fix waits for real completion, it
// does not skip the wait.
func TestWaitForFunctionWithTimeout_NoWaitGroup_StuckStillTimesOut(t *testing.T) {
	common.ResetGlobalState()
	if _, err := app.NewAppManager("stuck-app").CreateApp(); err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}
	lm := local.NewLocalManager("stuck-app", "stuck-local")
	if _, err := lm.CreateLocal("stuck-local"); err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}

	stuck := make(chan struct{}) // never closed: the goroutine ignores its context
	if err := lm.Go("stuck-nowg", func(ctx context.Context) error {
		<-stuck
		return nil
	}); err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	start := time.Now()
	if lm.WaitForFunctionWithTimeout("stuck-nowg", 200*time.Millisecond) {
		t.Fatal("stuck goroutine reported as completed")
	}
	if took := time.Since(start); took < 150*time.Millisecond {
		t.Fatalf("timeout returned after %v, before the 200ms limit", took)
	}
}
