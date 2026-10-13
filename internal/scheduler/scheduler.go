package scheduler

import (
	"context"
	"sync"
	"time"
	"uptime/internal/api"
	"uptime/internal/monitor"
)

func RunScheduler(ctx context.Context, store *monitor.MonitorStore) {
	started := make(map[int]bool)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var wg sync.WaitGroup

	for {
		select {
		case <-ticker.C:
			for id, m := range store.List() {
				if _, ok := started[id]; !ok {
					started[id] = true
					wg.Add(1)
					go func(m monitor.Monitor) {
						defer wg.Done()
						api.RunMonitor(ctx, m)
					}(m)
				}
			}
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}

}
