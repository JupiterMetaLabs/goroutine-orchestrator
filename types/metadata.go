package types

import (
	"sync"
	"time"
)

// This file is to set the metadata for the project
func (GM *GlobalManager) NewMetadata() *Metadata {
	// If metadata already exists, return it
	if GM.GetMetadata() != nil {
		return GM.GetMetadata()
	}

	md := &Metadata{
		metadataMu:              &sync.RWMutex{},
		Metrics:         false,
		ShutdownTimeout: 10 * time.Second,
		UpdateInterval:  UpdateInterval,
	}
	GM.SetMetadata(md)
	return md
}

// Right now our codebase not respecting the max routines limit, will set it
func (MD *Metadata) SetMaxRoutines(maxroutines int) *Metadata {
	// Lock and update
	MD.metadataMu.Lock()
	defer MD.metadataMu.Unlock()
	MD.MaxRoutines = maxroutines
	return MD
}

func (MD *Metadata) SetShutdownTimeout(timeout time.Duration) *Metadata {
	// Lock and update
	MD.metadataMu.Lock()
	defer MD.metadataMu.Unlock()
	MD.ShutdownTimeout = timeout
	// Set to global variable
	ShutdownTimeout = timeout
	return MD
}

func (MD *Metadata) SetMetrics(metrics bool, URL string, interval time.Duration) *Metadata {
	// Lock and update
	MD.metadataMu.Lock()
	defer MD.metadataMu.Unlock()
	MD.Metrics = metrics
	MD.MetricsURL = URL
	MD.UpdateInterval = interval
	// Set to global variable (similar to ShutdownTimeout)
	// The collector will observe this change via types.UpdateInterval
	if interval > 0 {
		UpdateInterval = interval
	}
	return MD
}

func (MD *Metadata) GetMetadata() *Metadata {
	// Lock and update
	MD.metadataMu.RLock()
	defer MD.metadataMu.RUnlock()
	return MD
}

func (MD *Metadata) UpdateIntervalTime(time time.Duration) time.Duration {
	// Lock and update
	MD.metadataMu.Lock()
	defer MD.metadataMu.Unlock()
	// Set to global variable
	if time > 0 {
		UpdateInterval = time
	}
	return MD.UpdateInterval
}

// ✅ ADD these getter methods
func (MD *Metadata) GetMetrics() bool {
    MD.metadataMu.RLock()
    defer MD.metadataMu.RUnlock()
    return MD.Metrics
}

func (MD *Metadata) GetMaxRoutines() int {
    MD.metadataMu.RLock()
    defer MD.metadataMu.RUnlock()
    return MD.MaxRoutines
}

func (MD *Metadata) GetUpdateInterval() time.Duration {
    MD.metadataMu.RLock()
    defer MD.metadataMu.RUnlock()
    return MD.UpdateInterval
}

func (MD *Metadata) GetShutdownTimeout() time.Duration {
    MD.metadataMu.RLock()
    defer MD.metadataMu.RUnlock()
    return MD.ShutdownTimeout
}

func (MD *Metadata) GetMetricsURL() string {
    MD.metadataMu.RLock()
    defer MD.metadataMu.RUnlock()
    return MD.MetricsURL
}