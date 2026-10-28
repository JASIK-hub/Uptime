package api

import (
	"bytes"
	"context"
	"errors"
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

func TestCreateMonitorTable(t *testing.T) {
	tt := []struct {
		name           string
		body           string
		mockAddFn      func(ctx context.Context, m monitor.Monitor) (int64, error)
		expectedStatus int
	}{
		{
			name: "successfull creation",
			body: `{"url":"https://example.com","intervalSeconds": 10,"timeoutSeconds":5}`,
			mockAddFn: func(ctx context.Context, m monitor.Monitor) (int64, error) {
				return 999, nil
			},
			expectedStatus: http.StatusCreated,
		}, {
			name: "Ошибка базы данных",
			body: `{"url":"https://example.com","intervalSeconds": 10,"timeoutSeconds":5}`,
			mockAddFn: func(ctx context.Context, m monitor.Monitor) (int64, error) {
				return 0, errors.New("db connection lost")

			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Битый JSON в запросе",
			body:           `not-a-json-string`,
			mockAddFn:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid URL format",
			body:           `{"url":"/example.com","intervalSeconds": 10,"timeoutSeconds":5}`,
			mockAddFn:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "negative interval time",
			body:           `{"url":"https://example.com","intervalSeconds": -1,"timeoutSeconds":5}`,
			mockAddFn:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "negative timeout",
			body:           `{"url":"https://example.com","intervalSeconds": 10,"timeoutSeconds":-1}`,
			mockAddFn:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/monitor", bytes.NewReader([]byte(tc.body)))
			rec := httptest.NewRecorder()

			handler := NewMonitorHandler(&MockMonitorRepository{AddFn: tc.mockAddFn})
			handler.CreateMonitor(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d:", tc.expectedStatus, rec.Code)
			}
		})
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
