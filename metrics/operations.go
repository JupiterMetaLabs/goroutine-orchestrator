package metrics

import (
	"time"
)

// RecordGoroutineOperation records a goroutine operation
func RecordGoroutineOperation(operation, appName, localName, functionName string) {
	if !IsMetricsEnabled() {
		return
	}
	GoroutineOperationsTotal.WithLabelValues(operation, appName, localName, functionName).Inc()
}

// RecordManagerOperation records a manager operation
func RecordManagerOperation(managerType, operation, appName string) {
	if !IsMetricsEnabled() {
		return
	}
	ManagerOperationsTotal.WithLabelValues(managerType, operation, appName).Inc()
}

// RecordFunctionOperation records a function operation
func RecordFunctionOperation(operation, appName, localName, functionName string) {
	if !IsMetricsEnabled() {
		return
	}
	FunctionOperationsTotal.WithLabelValues(operation, appName, localName, functionName).Inc()
}

// RecordOperationError records an operation error
func RecordOperationError(operationType, operation, errorType string) {
	if !IsMetricsEnabled() {
		return
	}
	OperationErrorsTotal.WithLabelValues(operationType, operation, errorType).Inc()
}

// RecordGoroutineOperationDuration records the duration of a goroutine operation
func RecordGoroutineOperationDuration(operation string, duration time.Duration, appName, localName, functionName string) {
	if !IsMetricsEnabled() {
		return
	}
	GoroutineOperationDuration.WithLabelValues(operation, appName, localName, functionName).Observe(duration.Seconds())
}

// RecordManagerOperationDuration records the duration of a manager operation
func RecordManagerOperationDuration(managerType, operation string, duration time.Duration, appName string) {
	if !IsMetricsEnabled() {
		return
	}
	ManagerOperationDuration.WithLabelValues(managerType, operation, appName).Observe(duration.Seconds())
}

// RecordShutdownDuration records the duration of a shutdown operation
func RecordShutdownDuration(managerType, shutdownType string, duration time.Duration, appName, localName string) {
	if !IsMetricsEnabled() {
		return
	}
	ShutdownDuration.WithLabelValues(managerType, shutdownType, appName, localName).Observe(duration.Seconds())
}

// RecordShutdownGoroutinesRemaining records the number of goroutines remaining after shutdown
func RecordShutdownGoroutinesRemaining(managerType, appName, localName string, count int) {
	if !IsMetricsEnabled() {
		return
	}
	ShutdownGoroutinesRemaining.WithLabelValues(managerType, appName, localName).Set(float64(count))
}
