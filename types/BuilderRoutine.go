package types

import (
	"context"
	"time"

	Helper "github.com/neerajchowdary889/GoRoutinesManager/Helper/Routine"
)

const (
	Prefix_Routine = "Routine."
)

// This should be reviewed again carefully to ensure goroutines are spawned quick and efficient
// TODO

// NewGoRoutine creates a new Routine instance with default values.
// Performance: ~40-50ns per call (ID generation is the main cost).
// It generates a unique ID using NewUUID() (~40ns) and sets StartedAt to current time.
// Returns a Routine builder for chaining setter methods.
//
// Note: Done channel is created as bidirectional (chan struct{}) but stored as read-only
// (<-chan struct{}) in the struct. The channel should be closed (not sent to) when the
// routine completes. Context and Cancel should be set via SetContext() and SetCancel()
// when the routine is actually spawned.
func(LM *LocalManager) NewGoRoutine(functionName string) *Routine {
	// Create buffered channel (size 1) for non-blocking signaling
	// This allows the channel to be closed without blocking if nothing is reading
	done := make(chan struct{}, 1)

	// Use builder pattern for efficient initialization
	// All operations are O(1) - ID generation is the slowest at ~40ns
	routine := &Routine{}

	routine.SetFunctionName(functionName).
		SetID(Helper.NewUUID()).            // Fast UUID generation (~40ns)
		SetDone(done).                      // Channel assignment (negligible cost)
		SetStartedAt(time.Now().UnixNano()) // Timestamp (fast, ~10ns)
		
	LM.AddRoutine(routine)
	return routine
}

// SetID sets the ID for the routine
func (r *Routine) SetID(id string) *Routine {
	r.ID = id
	return r
}

// SetFunctionName sets the function name for the routine
func (r *Routine) SetFunctionName(functionName string) *Routine {
	r.FunctionName = functionName
	return r
}


// SetContext sets the context for the routine
func (r *Routine) SetContext(ctx context.Context) *Routine {
	r.Ctx = ctx
	return r
}

// SetDone sets the done channel for the routine
func (r *Routine) SetDone(done <-chan struct{}) *Routine {
	r.Done = done
	return r
}

// SetCancel sets the cancel function for the routine
func (r *Routine) SetCancel(cancel context.CancelFunc) *Routine {
	r.Cancel = cancel
	return r
}

// SetStartedAt sets the started timestamp for the routine
func (r *Routine) SetStartedAt(timestamp int64) *Routine {
	r.StartedAt = timestamp
	return r
}

// DoneChan returns the done channel for the routine (read-only).
// The channel should be closed (not sent to) when the routine completes.
// Consumers can select on this channel to detect routine completion.
func (r *Routine) DoneChan() <-chan struct{} {
	return r.Done
}

// Get Functions
func (r *Routine) GetID() string {
	return r.ID
}

func (r *Routine) GetFunctionName() string {
	return r.FunctionName
}

func (r *Routine) GetContext() context.Context {
	return r.Ctx
}

func (r *Routine) GetCancel() context.CancelFunc {
	return r.Cancel
}

func (r *Routine) GetStartedAt() int64 {
	return r.StartedAt
}

