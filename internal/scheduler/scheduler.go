package scheduler

import (
	"context"
	"log"
	"sync"
	"time"
	"uptime/internal/monitor"
)

type Lister interface {
	List(ctx context.Context) ([]monitor.Monitor, error)
}

type Schedule struct {
	lister    Lister
	running   map[int64]bool
	nextCheck map[int64]time.Time
	check     func(context.Context, monitor.Monitor)
}

type checkFinished struct {
	monitorId  int64
	finishedAt time.Time
	interval   time.Duration
}

func NewSchedule(l Lister) *Schedule {
	return &Schedule{
		lister:    l,
		check:     executeCheck,
		nextCheck: make(map[int64]time.Time),
		running:   make(map[int64]bool),
	}
}

func RunMonitor(ctx context.Context, m monitor.Monitor) error {
	ticker := time.NewTicker(m.Interval())
	defer ticker.Stop()

	executeCheck(ctx, m)

	for {
		select {
		case <-ticker.C:
			if ctx.Err() != nil {
				return ctx.Err()
			}
			executeCheck(ctx, m)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func executeCheck(ctx context.Context, m monitor.Monitor) {
	c, err := monitor.HTTPCheck(ctx, m)
	if err != nil {
		log.Printf("check %s failed: %v", m.URL(), err)
		return
	}
	log.Printf("check %s: success=%t, duration=%v",
		m.URL(), c.Success, c.Duration)
}

func (s *Schedule) Run(ctx context.Context) error {
	timer := time.NewTimer(0)
	defer timer.Stop()

	var wg sync.WaitGroup
	jobs := make(chan monitor.Monitor)
	finished := make(chan checkFinished)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case m := <-jobs:
					if ctx.Err() != nil {
						return
					}
					s.check(ctx, m)

					select {
					case finished <- checkFinished{
						monitorId:  m.ID(),
						finishedAt: time.Now(),
						interval:   m.Interval(),
					}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	defer wg.Wait()

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
				if s.running[id] {
					continue
				}
				nextTime, exists := s.nextCheck[id]
				if exists && time.Now().Before(nextTime) {
					continue
				}
				select {
				case jobs <- m:
					s.running[id] = true
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
			}
		case done := <-finished:
			s.running[done.monitorId] = false
			s.nextCheck[done.monitorId] = done.finishedAt.Add(done.interval)

		case <-ctx.Done():
			return ctx.Err()
		}
	}

}
