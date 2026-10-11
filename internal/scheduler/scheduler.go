package scheduler

import (
	"context"
	"time"
	"uptime/internal/api"
	"uptime/internal/monitor"
)

func RunScheduler(ctx context.Context, store *monitor.MonitorStore) {
	started := make(map[int]bool)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for id, m := range store.List() {
				if _, ok := started[id]; !ok {
					started[id] = true
					go api.RunMonitor(ctx, m)
				}
			}
		case <-ctx.Done():
			return
		}
	}

}
