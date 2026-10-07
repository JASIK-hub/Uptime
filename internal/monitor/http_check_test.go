package monitor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPCheck(t *testing.T) {
	okStatus := http.StatusOK
	serverError := http.StatusInternalServerError
	badRequest := http.StatusBadRequest
	tt := map[string]struct {
		statusCode int
		want       CheckResult
	}{
		"200 statusCode": {
			statusCode: http.StatusOK,
			want: CheckResult{
				Success:    true,
				StatusCode: &okStatus,
			},
		},
		"500 statusCode": {
			statusCode: serverError,
			want: CheckResult{
				Success:    false,
				StatusCode: &serverError,
			},
		},
		"400 statusCode": {
			statusCode: badRequest,
			want: CheckResult{
				Success:    false,
				StatusCode: &badRequest,
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer ts.Close()

			monitor, err := NewMonitor(ts.URL, time.Second)
			if err != nil {
				t.Fatalf("create monitor: %v", err)
			}
			got, err := HTTPCheck(context.Background(), monitor)
			if err != nil {
				t.Fatalf("failed to check http: %v", err)
			}
			mustCheckResult(t, got, tc.want)
		})
	}
}

func mustCheckResult(t *testing.T, got, want CheckResult) {
	t.Helper()

	if got.StatusCode == nil || want.StatusCode == nil {
		t.Fatal("expected non-nil status codes")
	}

	if *got.StatusCode != *want.StatusCode {
		t.Fatalf("expected %v, got %v", *want.StatusCode, *got.StatusCode)
	}

	if got.Success != want.Success {
		t.Fatalf("expected %v, got %v", want.Success, got.Success)
	}
}

func TestHTTP_Timeout(t *testing.T) {
	t.Run("timeout exceeded", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(60 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()
		monitor, err := NewMonitor(ts.URL, 20*time.Millisecond)
		if err != nil {
			t.Fatalf("failed to create monitor: %v", err)
		}
		got, err := HTTPCheck(context.Background(), monitor)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected timeout, got %v", err)
		}
		if got.Duration <= 0 {
			t.Errorf("expected duration to be more than 0: %v", err.Error())
		}
		if got.StartTime.IsZero() {
			t.Errorf("expected start time to be more than 0: %v", err.Error())
		}
		if got.StatusCode != nil {
			t.Errorf("expected status code to be empty: %v", err.Error())
		}
		if got.Success {
			t.Errorf("expected Success=false, got true: %v", err.Error())
		}

	})
}

func TestHTTPCheck_CanceledContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	monitor, err := NewMonitor(ts.URL, time.Second)
	if err != nil {
		t.Fatalf("failed to create monitor: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := HTTPCheck(ctx, monitor)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context to be cancelled, got :%v", err)
	}
	if got.Duration < 0 {
		t.Errorf("expected non-negative Duration: %v", err.Error())
	}
	if got.StartTime.IsZero() {
		t.Errorf("expected non-zero StartTime: %v", err.Error())
	}
	if got.StatusCode != nil {
		t.Errorf("expected status code to be empty: %v", err.Error())
	}
	if got.Success {
		t.Errorf("expected Success=false, got true: %v", err.Error())
	}
}
