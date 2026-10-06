package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHttpCheck(t *testing.T) {
	okStatus := http.StatusOK
	serverError := http.StatusInternalServerError
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
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer ts.Close()

			monitor, err := NewMonitor(ts.URL, 5*time.Second)
			if err != nil {
				t.Fatalf("create monitor: %v", err)
			}
			got, err := HttpCheck(monitor)
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
