package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"
	"uptime/internal/monitor"
)

type MockLister struct {
	Monitor   []monitor.Monitor
	callcount int
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
