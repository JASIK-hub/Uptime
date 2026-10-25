package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"uptime/internal/monitor"
)

type MockMonitorRepository struct {
	ListFn func(ctx context.Context) ([]monitor.Monitor, error)
	AddFn  func(ctx context.Context, m monitor.Monitor) (int64, error)
}

func (m *MockMonitorRepository) List(ctx context.Context) ([]monitor.Monitor, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}

func (m *MockMonitorRepository) Add(ctx context.Context, mon monitor.Monitor) (int64, error) {
	if m.AddFn != nil {
		return m.AddFn(ctx, mon)
	}
	return 0, nil
}

func TestCreateMonitor(t *testing.T) {
	body := `{"url":"https://example.com","intervalSeconds": 10,"timeoutSeconds":5}`
	req := httptest.NewRequest(http.MethodPost, "/monitor", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	mockRepo := &MockMonitorRepository{
		AddFn: func(ctx context.Context, m monitor.Monitor) (int64, error) {
			return 999, nil
		},
	}

	rec := httptest.NewRecorder()
	handler := NewMonitorHandler(mockRepo)
	handler.CreateMonitor(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestListMonitors(t *testing.T) {
	mockRepo := &MockMonitorRepository{
		ListFn: func(ctx context.Context) ([]monitor.Monitor, error) {
			return nil, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/monitors", nil)
	req.Header.Set("Content-Type", "application/json")
	handler := NewMonitorHandler(mockRepo)

	rec := httptest.NewRecorder()
	handler.ListMonitors(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
