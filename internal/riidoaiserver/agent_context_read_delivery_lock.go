package riidoaiserver

import (
	"context"
	"sync"
	"time"
)

// The publisher holds the store lock through snapshot persistence after fanout.
// Wait only inside the bounded metadata worker, never the SSE writer. A cancel
// or timeout must not leave a goroutine blocked on Mutex.Lock.
func lockContextReadDelivery(ctx context.Context, mu *sync.Mutex) bool {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for ctx.Err() == nil {
		if mu.TryLock() {
			if ctx.Err() == nil {
				return true
			}
			mu.Unlock()
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
		}
	}
	return false
}
