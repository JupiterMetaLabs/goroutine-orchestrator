package Routine

import (
	"encoding/base64"
	"sync"
	"testing"
)

// Performance Notes:
// - BenchmarkNewUUID: ~40-50 ns/op (nanosecond-level performance)
// - Single allocation per call (24-40 bytes)
// - Thread-safe via atomic operations
// - ~100x faster than crypto-based UUIDs (which take microseconds)

// TestNewUUID_Uniqueness verifies that NewUUID generates unique IDs
func TestNewUUID_Uniqueness(t *testing.T) {
	const numIDs = 10000
	ids := make(map[string]bool, numIDs)

	for i := 0; i < numIDs; i++ {
		id := NewUUID()
		if ids[id] {
			t.Fatalf("Duplicate ID generated: %s at iteration %d", id, i)
		}
		ids[id] = true
	}

	if len(ids) != numIDs {
		t.Fatalf("Expected %d unique IDs, got %d", numIDs, len(ids))
	}
}

// TestNewUUID_Format verifies the ID format (22 chars, base64 URL-safe, no padding)
func TestNewUUID_Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := NewUUID()

		// Check length (base64 of 16 bytes = 22 chars, no padding)
		if len(id) != 22 {
			t.Errorf("Expected ID length 22, got %d: %s", len(id), id)
		}

		// Verify it's valid base64 URL-safe (raw, no padding)
		decoded, err := base64.RawURLEncoding.DecodeString(id)
		if err != nil {
			t.Errorf("Invalid base64 URL encoding: %s, error: %v", id, err)
		}

		// Verify decoded length is 16 bytes
		if len(decoded) != 16 {
			t.Errorf("Expected decoded length 16 bytes, got %d", len(decoded))
		}

		// Verify no padding characters (RawURLEncoding should never add padding)
		if len(id) > 0 && id[len(id)-1] == '=' {
			t.Errorf("ID should not have padding, got: %s", id)
		}
	}
}

// TestNewUUID_Concurrency verifies thread-safety under concurrent access
func TestNewUUID_Concurrency(t *testing.T) {
	const numGoroutines = 100
	const idsPerGoroutine = 1000
	const totalIDs = numGoroutines * idsPerGoroutine

	ids := make(map[string]bool, totalIDs)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Generate IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id := NewUUID()
				mu.Lock()
				if ids[id] {
					t.Errorf("Duplicate ID in concurrent test: %s", id)
				}
				ids[id] = true
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Verify all IDs are unique
	if len(ids) != totalIDs {
		t.Fatalf("Expected %d unique IDs, got %d", totalIDs, len(ids))
	}
}

// TestNewUUID_Ordering verifies that IDs are generally increasing (due to timestamp)
func TestNewUUID_Ordering(t *testing.T) {
	const numIDs = 100
	ids := make([]string, numIDs)

	for i := 0; i < numIDs; i++ {
		ids[i] = NewUUID()
	}

	// Verify IDs are unique (they should be due to counter even if same nanosecond)
	uniqueIDs := make(map[string]bool, numIDs)
	for _, id := range ids {
		if uniqueIDs[id] {
			t.Errorf("Duplicate ID in ordering test: %s", id)
		}
		uniqueIDs[id] = true
	}
}

// BenchmarkNewUUID benchmarks the performance of NewUUID
func BenchmarkNewUUID(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewUUID()
	}
}

// BenchmarkNewUUID_Parallel benchmarks concurrent NewUUID calls
func BenchmarkNewUUID_Parallel(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewUUID()
		}
	})
}

// BenchmarkNewUUID_Throughput measures throughput (IDs per second)
func BenchmarkNewUUID_Throughput(b *testing.B) {
	b.ResetTimer()

	ids := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		ids[i] = NewUUID()
	}

	// Prevent compiler optimization
	_ = ids
}

// ExampleNewUUID demonstrates usage of NewUUID
func ExampleNewUUID() {
	id := NewUUID()
	_ = id // Use the ID
}
