package routine

import (
	"encoding/base64"
	"encoding/binary"
	"sync/atomic"
	"time"
)

var (
	// Atomic counter for unique IDs within the same nanosecond
	counter uint64
)

// NewUUID generates a fast, unique ID for goroutines.
// Format: base64(timestamp_nanoseconds + atomic_counter)
// This is ~100x faster than crypto-based UUIDs (~nanoseconds vs microseconds).
// The ID is unique within the process and suitable for goroutine tracking.
// Returns a 22-character base64 URL-safe string (no padding).
func NewUUID() string {
	// Get current time in nanoseconds
	now := time.Now().UnixNano()

	// Get next atomic counter value
	seq := atomic.AddUint64(&counter, 1)

	// Combine timestamp (8 bytes) + counter (8 bytes) = 16 bytes
	// Encode as base64 URL-safe (22 characters, no padding)
	id := make([]byte, 16)

	// Write timestamp (little-endian)
	binary.LittleEndian.PutUint64(id[0:8], uint64(now))

	// Write counter (little-endian)
	binary.LittleEndian.PutUint64(id[8:16], seq)

	return base64.RawURLEncoding.EncodeToString(id)
}
