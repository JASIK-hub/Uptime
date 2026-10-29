package scheduler

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"time"
	"uptime/internal/api"
	"uptime/internal/monitor"
)

type Lister interface {
	List(ctx context.Context) ([]monitor.Monitor, error)
}

type Schedule struct {
	lister         Lister
	activeMonitors map[int64]context.CancelFunc
	mu             sync.Mutex
}

func NewSchedule(l Lister) *Schedule {
	return &Schedule{
		lister:         l,
		activeMonitors: make(map[int64]context.CancelFunc),
	}
}

func (s *Schedule) Run(ctx context.Context) error {
	timer := time.NewTimer(0)
	defer timer.Stop()
	var wg sync.WaitGroup

	currentDelay := time.Second
	baseDelay := time.Second
	maxDelay := time.Minute

	defer func() {
		s.mu.Lock()
		for _, cancel := range s.activeMonitors {
			cancel()
		}
		s.mu.Unlock()
		wg.Wait()
	}()

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

			currentIds := make(map[int64]bool)
			for _, m := range monitors {
				currentIds[m.ID()] = true
			}

			s.mu.Lock()
			for id, cancel := range s.activeMonitors {
				if !currentIds[id] {
					cancel()
					delete(s.activeMonitors, id)
				}
			}

			for _, m := range monitors {
				id := m.ID()
				if _, exists := s.activeMonitors[id]; !exists {

					workerCtx, cancel := context.WithCancel(ctx)
					s.activeMonitors[id] = cancel

					wg.Add(1)
					go func(wCtx context.Context, m monitor.Monitor, workerId int64) {
						defer wg.Done()
						defer func() {
							s.mu.Lock()
							delete(s.activeMonitors, workerId)
							s.mu.Unlock()
						}()

						if err := api.RunMonitor(wCtx, m); err != nil {
							slog.Error(
								"monitor execution failed",
								"worker_id", workerId,
								"monitor_id", m.ID(),
								"error", err)
							return
						}
					}(workerCtx, m, id)
				}
			}
			s.mu.Unlock()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}
