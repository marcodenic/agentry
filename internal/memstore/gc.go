package memstore

import (
	"os"
	"strconv"
	"sync"
	"time"
)

var gcOnce sync.Once

// StartDefaultGC starts a background goroutine that periodically calls CleanupExpired()
// on the default SharedStore. It runs once per process.
func StartDefaultGC(interval time.Duration) {
	gcOnce.Do(func() {
		if env := os.Getenv("AGENTRY_STORE_GC_SEC"); env != "" {
			if sec, err := strconv.Atoi(env); err == nil && sec > 0 {
				interval = time.Duration(sec) * time.Second
			}
		}
		if interval <= 0 {
			interval = 60 * time.Second
		}
		go func() {
			t := time.NewTicker(interval)
			defer t.Stop()
			for range t.C {
				if Get() != nil {
					_ = Get().CleanupExpired()
				}
			}
		}()
	})
}
