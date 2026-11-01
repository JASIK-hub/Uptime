package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
	db "uptime/internal/infra/postgres/sqlc"
	"uptime/internal/monitor"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MonitorRepository struct {
	queries *db.Queries
}

func NewMonitorRepository(pool *pgxpool.Pool) *MonitorRepository {
	return &MonitorRepository{queries: db.New(pool)}
}

func (r *MonitorRepository) Add(ctx context.Context, m monitor.Monitor) (int64, error) {
	timeout, err := durationSeconds(m.Timeout())
	if err != nil {
		return 0, fmt.Errorf("timeout validation error: %w", err)
	}
	interval, err := durationSeconds(m.Interval())
	if err != nil {
		return 0, fmt.Errorf("interval validation error: %w", err)
	}
	id, err := r.queries.AddMonitor(
		ctx,
		db.AddMonitorParams{
			Url:             m.URL(),
			TimeoutSeconds:  timeout,
			IntervalSeconds: interval,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert monitor: %w", err)
	}
	return id, nil
}

func (r *MonitorRepository) List(ctx context.Context) ([]monitor.Monitor, error) {
	rows, err := r.queries.ListMonitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("obtain monitor list: %w", err)
	}
	monitors := make([]monitor.Monitor, len(rows))
	for i, row := range rows {
		m, err := monitor.RestoreMonitor(
			row.ID,
			row.Url,
			time.Duration(row.TimeoutSeconds)*time.Second,
			time.Duration(row.IntervalSeconds)*time.Second,
		)
		if err != nil {
			return nil, fmt.Errorf("restore monitor %d: %w", row.ID, err)
		}
		monitors[i] = m
	}
	return monitors, nil
}

func durationSeconds(d time.Duration) (int32, error) {
	if d <= 0 {
		return 0, errors.New("duration must be positive")
	}
	if d%time.Second != 0 {
		return 0, errors.New("durations must be whole number")
	}

	seconds := d / time.Second
	if seconds > math.MaxInt32 {
		return 0, errors.New("duration exceeds database integer range")
	}
	return int32(seconds), nil
}
