package utility

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

// Global counter to prevent collisions if called within the same nanosecond
var counter uint64

func GenerateUnique() string {
	// 1. Get the current time in nanoseconds
	now := time.Now().UnixNano()

	// 2. Increment the counter atomically (thread-safe)
	count := atomic.AddUint64(&counter, 1)

	// 3. Grab a few random bytes to secure it against prediction
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)

	// 4. Combine time, sequential count, and randomness into a unique string
	return fmt.Sprintf("%x-%x-%s", now, count, hex.EncodeToString(randomBytes))
}


