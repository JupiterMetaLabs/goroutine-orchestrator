package types

import (
	"time"
)

// This file is to set the metadata for the project
func (GM *GlobalManager) NewMetadata() *Metadata {
	// If metadata already exists, return it
	if GM.GetMetadata() != nil {
		return GM.GetMetadata()
	}

	md := &Metadata{
		Metrics:         false,
		ShutdownTimeout: 10 * time.Second,
	}
	GM.SetMetadata(md)
	return md
}

// Right now our codebase not respecting the max routines limit, will set it
func (MD *Metadata) SetMaxRoutines(maxroutines int) *Metadata {
	MD.MaxRoutines = maxroutines
	return MD
}

func (MD *Metadata) SetShutdownTimeout(timeout time.Duration) *Metadata {
	MD.ShutdownTimeout = timeout
	// Set to global variable
	ShutdownTimeout = timeout
	return MD
}

func (MD *Metadata) SetMetrics(metrics bool, URL string) *Metadata {
	MD.Metrics = metrics
	MD.MetricsURL = URL

	return MD
}

func (MD *Metadata) GetMetadata() *Metadata {
	return MD
}
