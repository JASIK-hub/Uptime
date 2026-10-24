package scheduler

import (
	"context"
	"log"
	"sync"
	"time"
	"uptime/internal/api"
	"uptime/internal/monitor"
)

type Lister interface {
	List(ctx context.Context) ([]monitor.Monitor, error)
}

type Schedule struct {
	lister Lister
}

func NewSchedule(l Lister) *Schedule {
	return &Schedule{
		lister: l,
	}
}

func (s Schedule) Run(ctx context.Context) {
	started := make(map[int64]bool)
	timer := time.NewTimer(0)
	defer timer.Stop()
	var wg sync.WaitGroup

	currentDelay := time.Second
	baseDelay := time.Second
	maxDelay := time.Minute
	for {
		select {
		case <-timer.C:
			monitors, err := s.lister.List(ctx)

			if err != nil {
				log.Printf("error: %v", err)

				currentDelay = currentDelay * 2

				if currentDelay > maxDelay {
					currentDelay = maxDelay
				}
				timer.Reset(currentDelay)
				continue
			}
			currentDelay = baseDelay
			timer.Reset(currentDelay)

			for _, m := range monitors {
				id := m.ID()
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
