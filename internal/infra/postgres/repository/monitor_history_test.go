package repository

import (
	"context"
	"net/http"
	"testing"
	"time"
	"uptime/internal/monitor"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestMonitorRepostiry_SaveCheck(t *testing.T) {

	pool := newPool(t)
	r := NewMonitorRepository(pool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, err := monitor.NewMonitor("https://example.com/health",
		5*time.Second,
		30*time.Second)

	if err != nil {
		t.Fatalf("could not create monitor: %v", err)
	}

	monitor_id, err := r.Add(ctx, m)

	if err != nil {
		t.Fatalf("could not add monitor: %v", err)
	}

	statusCode := http.StatusUnauthorized
	result := monitor.CheckResult{
		StartTime:  time.Now(),
		Duration:   800 * time.Microsecond,
		Success:    false,
		StatusCode: &statusCode,
	}

	if err = r.SaveCheck(ctx, monitor_id, result, nil); err != nil {
		t.Fatalf("save check: %v", err)
	}

	var (
		success       bool
		duration_ms   int32
		status_code   pgtype.Int4
		error_message pgtype.Text
	)

	err = pool.QueryRow(ctx,
		`SELECT success,duration_ms,status_code,error_message
	     FROM monitors_history 
		 WHERE monitor_id = $1`, monitor_id).Scan(&success, &duration_ms, &status_code, &error_message)

	if err != nil {
		t.Fatalf("read saved check: %v", err)
	}
	if success {
		t.Errorf("expected success to be: %v, got: %v", result.Success, success)
	}
	if status_code.Int32 != http.StatusUnauthorized {
		t.Errorf("expected status_code to be: %d, got: %d", *result.StatusCode, statusCode)
	}
	if error_message.Valid {
		t.Errorf("want error_message=NULL, got %q", error_message.String)
	}
}
