package types

import (
	"context"
	"time"

	Helper "github.com/neerajchowdary889/GoRoutinesManager/Helper/Routine"
)

const (
	Prefix_Routine = "Routine."
)

// NewGoRoutine creates a new Routine instance with default values.
// It generates a unique ID using NewUUID() and sets StartedAt to current time.
// Returns a Routine builder for chaining setter methods.
// Note: Done channel is created as bidirectional but stored as read-only in the struct.
func NewGoRoutine(functionName string) *Routine {
	// Create bidirectional channel - will be assigned to read-only field
	done := make(chan struct{}, 1)
	routine := &Routine{}

	routine.SetFunctionName(functionName).
		SetID(Helper.NewUUID()).
		SetDone(done). // Go allows bidirectional -> read-only assignment
		SetStartedAt(time.Now().UnixNano())

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

// DoneChan returns the done channel for the routine
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
