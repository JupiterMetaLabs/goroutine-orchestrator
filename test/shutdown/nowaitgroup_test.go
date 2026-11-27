package shutdown_test

import (
	"context"
	"testing"
	"time"

	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/app"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/global"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/interfaces"
	"github.com/JupiterMetaLabs/goroutine-orchestrator/manager/local"
	common "github.com/JupiterMetaLabs/goroutine-orchestrator/test/common"
)

// newNoWGLocal sets up an app and a local manager for the tests below.
func newNoWGLocal(t *testing.T, appName, localName string) interfaces.LocalGoroutineManagerInterface {
	t.Helper()
	common.ResetGlobalState()
	if _, err := app.NewAppManager(appName).CreateApp(); err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}
	lm := local.NewLocalManager(appName, localName)
	if _, err := lm.CreateLocal(localName); err != nil {
		t.Fatalf("CreateLocal() failed: %v", err)
	}
	return lm
}

// ctxWorker runs until its context is cancelled, like most real workers.
func ctxWorker(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// The bug: goroutines started WITHOUT AddToWaitGroup that stop as soon as
// they are cancelled made the safe shutdown wait its full timeout (10s),
// because WaitForFunction returned "no wait group" and
// WaitForFunctionWithTimeout never reported completion.
func TestSafeShutdown_NoWaitGroup_DoesNotWaitFullTimeout(t *testing.T) {
	lm := newNoWGLocal(t, "nowg-app", "nowg-local")
	for i := 0; i < 3; i++ {
		if err := lm.Go("nowg-worker", ctxWorker); err != nil {
			t.Fatalf("Go() failed: %v", err)
		}
	}
	if err := lm.Go("other-nowg-worker", ctxWorker); err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	start := time.Now()
	if err := lm.Shutdown(true); err != nil {
		t.Fatalf("Shutdown(true) failed: %v", err)
	}
	if took := time.Since(start); took > 2*time.Second {
		t.Fatalf("safe shutdown took %v; goroutines without a wait group must not cost the full timeout", took)
	}
}

// The same through the global manager, which is how applications call it.
func TestGlobalSafeShutdown_NoWaitGroup_IsFast(t *testing.T) {
	lm := newNoWGLocal(t, "nowg-global-app", "nowg-global-local")
	for i := 0; i < 3; i++ {
		if err := lm.Go("nowg-worker", ctxWorker); err != nil {
			t.Fatalf("Go() failed: %v", err)
		}
	}

	start := time.Now()
	if err := global.NewGlobalManager().Shutdown(true); err != nil {
		t.Fatalf("global Shutdown(true) failed: %v", err)
	}
	if took := time.Since(start); took > 2*time.Second {
		t.Fatalf("global safe shutdown took %v; want well under the 10s timeout", took)
	}
}

// WaitForFunction must really wait for goroutines without a wait group, not
// return early: it returns only once they have finished.
func TestWaitForFunction_NoWaitGroup_WaitsForCompletion(t *testing.T) {
	lm := newNoWGLocal(t, "nowg-wait-app", "nowg-wait-local")
	release := make(chan struct{})
	finished := make(chan struct{})
	if err := lm.Go("slow-nowg", func(ctx context.Context) error {
		<-release
		close(finished)
		return nil
	}); err != nil {
		t.Fatalf("Go() failed: %v", err)
	}

	waited := make(chan error, 1)
	go func() { waited <- lm.WaitForFunction("slow-nowg") }()

	select {
	case <-waited:
		t.Fatal("WaitForFunction returned while the goroutine was still running")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	select {
	case err := <-waited:
		if err != nil {
			t.Fatalf("WaitForFunction() = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForFunction did not return after the goroutine finished")
	}
	select {
	case <-finished:
	default:
		t.Fatal("WaitForFunction returned before the goroutine finished")
	}
}

// WaitForFunctionWithTimeout reports completion for a goroutine without a
// wait group as soon as it finishes. (The stuck case is in
// test/nowaitgroupstuck: it leaves a background waiter running, which must
// not overlap with a later ResetGlobalState in this package.)
func TestWaitForFunctionWithTimeout_NoWaitGroup(t *testing.T) {
	lm := newNoWGLocal(t, "nowg-timeout-app", "nowg-timeout-local")

	if err := lm.Go("quick-nowg", func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("Go() failed: %v", err)
	}
	if !lm.WaitForFunctionWithTimeout("quick-nowg", 2*time.Second) {
		t.Fatal("finished goroutine without a wait group reported as not completed")
	}
}

// A function with no goroutines at all has nothing to wait for.
func TestWaitForFunction_UnknownFunctionReturnsNil(t *testing.T) {
	lm := newNoWGLocal(t, "nowg-unknown-app", "nowg-unknown-local")
	if err := lm.WaitForFunction("never-started"); err != nil {
		t.Fatalf("WaitForFunction(unknown) = %v, want nil", err)
	}
}
