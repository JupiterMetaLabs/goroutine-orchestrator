package Local

import (
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/Manager/Interface"
)

// Option is a function that configures goroutine options.
// It can be used with Go() method to configure timeout, panic recovery, and wait group behavior.
// Option satisfies the Interface.GoroutineOption interface.
type Option func(*goroutineOptions)

// Ensure Option satisfies Interface.GoroutineOption
var _ Interface.GoroutineOption = Option(nil)

// goroutineOptions holds configuration for spawning goroutines
type goroutineOptions struct {
	timeout       *time.Duration // nil means no timeout
	panicRecovery bool           // whether to recover from panics
	waitGroupName string         // function name for wait group (empty means no wait group)
}

// defaultGoroutineOptions returns the default options
func defaultGoroutineOptions() *goroutineOptions {
	return &goroutineOptions{
		timeout:       nil,
		panicRecovery: true, // Enabled by default for production safety
		waitGroupName: "",
	}
}

// WithTimeout sets a timeout for the goroutine.
// When the timeout expires, the context will be cancelled automatically.
// The worker function should check ctx.Done() to handle timeout gracefully.
func WithTimeout(timeout time.Duration) Option {
	return func(opts *goroutineOptions) {
		opts.timeout = &timeout
	}
}

// WithPanicRecovery enables or disables panic recovery for the goroutine.
// Panic recovery is enabled by default for production safety.
// When enabled, panics in the worker function will be recovered,
// logged via metrics, and the goroutine will complete normally with cleanup.
// Set to false only if you want panics to crash the goroutine (not recommended).
func WithPanicRecovery(enabled bool) Option {
	return func(opts *goroutineOptions) {
		opts.panicRecovery = enabled
	}
}

// AddToWaitGroup adds the goroutine to a function wait group.
// The functionName parameter specifies which function wait group to use.
// The wait group will be created if it doesn't exist, and the goroutine
// will automatically call wg.Add(1) before spawning and wg.Done() on completion.
//
// Example:
//
//	localMgr.Go("worker", func(ctx context.Context) error { ... },
//	    AddToWaitGroup("worker"))
func AddToWaitGroup(functionName string) Option {
	return func(opts *goroutineOptions) {
		opts.waitGroupName = functionName
	}
}
