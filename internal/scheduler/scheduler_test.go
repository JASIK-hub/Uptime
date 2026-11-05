package scheduler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
	"uptime/internal/monitor"
)

type MockLister struct {
	Monitor   []monitor.Monitor
	callcount int
	check     func(context.Context, monitor.Monitor)
	mockErr   error
}

func (m *MockLister) List(ctx context.Context) ([]monitor.Monitor, error) {
	m.callcount++
	if m.mockErr != nil {
		return nil, m.mockErr
	}
	return m.Monitor, nil
}

func TestSchedule_Run_ExponentialBackoff(t *testing.T) {
	mockErr := errors.New("database connection failed")
	mockLister := &MockLister{mockErr: mockErr}

	scheduler := NewSchedule(mockLister)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		err := scheduler.Run(ctx)

		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			t.Errorf("unexpected scheduler error: %v", err)
		}
	}()
	<-done

	if mockLister.callcount == 0 {
		t.Errorf("expected at least one call to List, got 0")
	}
}

func TestSchedule_Run_Cancel(t *testing.T) {

	m, err := monitor.RestoreMonitor(
		1,
		"https://example.com",
		time.Second,
		30*time.Second,
	)
	if err != nil {
		t.Fatalf("restore monitor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{})
	stopped := make(chan struct{})
	result := make(chan error, 1)

	s := NewSchedule(&MockLister{
		Monitor: []monitor.Monitor{m},
	})

	s.check = func(ctx context.Context, m monitor.Monitor) {
		close(started)
		<-ctx.Done()
		close(stopped)
	}

	go func() {
		result <- s.Run(ctx)
	}()

	select {
	case <-started:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("check did not start")
	}

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not stop")
	}

	select {
	case <-stopped:
	default:
		t.Error("Run returned while check was still working")
	}

}

func TestSchedule_Run_MonitorLimit(t *testing.T) {
	monitors := make([]monitor.Monitor, 0, 11)

	for id := int64(1); id <= 11; id++ {
		m, err := monitor.RestoreMonitor(
			id,
			fmt.Sprintf("https://example.com/site-%d", id),
			5*time.Second,
			30*time.Second,
		)
		if err != nil {
			t.Fatalf("restore monitor %d: %v", id, err)
		}

		monitors = append(monitors, m)
	}

	s := NewSchedule(&MockLister{
		Monitor: monitors,
	})

	started := make(chan int64, 11)
	result := make(chan error, 1)
	release := make(chan struct{})

	seen := make(map[int64]bool)
	s.check = func(ctx context.Context, m monitor.Monitor) {
		started <- m.ID()
		select {
		case <-release:
		case <-ctx.Done():
			return
		}

	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		result <- s.Run(ctx)
	}()

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for i := 0; i < 10; i++ {
		select {
		case id := <-started:
			if seen[id] {
				t.Fatalf("monitor %d started twice", id)
			}
			seen[id] = true
		case <-timeout.C:
			t.Fatalf("expected 10 checks to start, got %d", i)
		}
	}

	select {
	case id := <-started:
		t.Fatalf("check %d started while 10 checks were still running", id)
	case <-time.After(1500 * time.Millisecond):
	}

	select {
	case release <- struct{}{}:
	case <-time.After(5 * time.Second):
		t.Fatal("could not release a check")
	}

	select {
	case id := <-started:
		if seen[id] {
			t.Fatalf(
				"monitor %d restarted instead of the remaining monitor",
				id,
			)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("remaining monitor did not start")
	}

	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not stop")
	}
}

func TestSchedule_Run_MonitorInterval(t *testing.T) {
	m, err := monitor.NewMonitor(
		"https://example.com",
		5*time.Second,
		4*time.Second)

	if err != nil {
		t.Fatalf("montor initialise error: %v", err)
	}
	s := NewSchedule(&MockLister{
		Monitor: []monitor.Monitor{m},
	})
	var first, second time.Time
	started := make(chan time.Time, 10)
	s.check = func(ctx context.Context, m monitor.Monitor) {
		started <- time.Now()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)

	go func() {
		result <- s.Run(ctx)
	}()

	select {
	case first = <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first check did not start")
	}

	select {
	case second = <-started:
	case <-time.After(7 * time.Second):
		t.Fatal("first check did not start")
	}
	interval := second.Sub(first)
	if interval < m.Interval() {
		t.Fatalf("check restarted too early, want at least: %v, got %v:", m.Interval(), interval)
	}
	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not stop")
	}
}
