package session

import (
	"context"
	"time"

	"github.com/marcodenic/agentry/pkg/memstore"
)

const historyBucket = "history"

// Start launches a background goroutine that periodically removes
// expired sessions from the store based on ttl. The cleanup runs on
// the provided interval until ctx is canceled.
func Start(ctx context.Context, store memstore.Cleaner, ttl, interval time.Duration) {
	if store == nil || ttl <= 0 || interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = store.Cleanup(context.Background(), historyBucket, ttl)
			}
		}
	}()
}
