package repository

import (
	"context"
	"fmt"
	"time"
	"uptime/internal/monitor"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MonitorRepository struct {
	pool *pgxpool.Pool
}

func NewMonitorRepository(pool *pgxpool.Pool) *MonitorRepository {
	return &MonitorRepository{pool: pool}
}

func (r *MonitorRepository) Add(ctx context.Context, m monitor.Monitor) (int64, error) {
	var id int64
	query := `INSERT INTO monitors (url,timeout_seconds,interval_seconds) VALUES ($1,$2,$3) RETURNING id;`
	row := r.pool.QueryRow(
		ctx,
		query,
		m.URL(),
		int64(m.Timeout()/time.Second),
		int64(m.Interval()/time.Second),
	)
	err := row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert monitor: %w", err)
	}
	return id, nil
}

func (r *MonitorRepository) List(ctx context.Context) ([]monitor.Monitor, error) {
	query := "SELECT id,url,timeout_seconds,interval_seconds FROM monitors ORDER BY id;"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("obtain monitor list: %w", err)
	}
	defer rows.Close()
	var monitors []monitor.Monitor
	for rows.Next() {
		var (
			id       int64
			rawUrl   string
			timeout  int
			interval int
		)
		if err := rows.Scan(&id, &rawUrl, &timeout, &interval); err != nil {
			return nil, fmt.Errorf("scan monitor: %w", err)
		}
		m, err := monitor.RestoreMonitor(id, rawUrl, time.Duration(timeout)*time.Second, time.Duration(interval)*time.Second)
		if err != nil {
			return nil, fmt.Errorf("error reading monitors: %w", err)
		}
		monitors = append(monitors, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate monitors: %w", err)
	}
	return monitors, nil
}
