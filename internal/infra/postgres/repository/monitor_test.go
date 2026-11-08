package repository

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"
	"uptime/internal/monitor"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	config, err := pgxpool.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatalf("parse database config: %v", err)
	}
	port, err := strconv.ParseUint(os.Getenv("POSTGRES_TEST_PORT"), 10, 16)

	if err != nil || port == 0 {
		t.Fatalf("POSTGRES_TEST_PORT must be a valid non-zero port")
	}

	config.ConnConfig.Host = os.Getenv("POSTGRES_TEST_HOST")
	config.ConnConfig.Port = uint16(port)
	config.ConnConfig.Database = os.Getenv("POSTGRES_TEST_DB")
	config.ConnConfig.Password = os.Getenv("POSTGRES_TEST_PASSWORD")
	config.ConnConfig.User = os.Getenv("POSTGRES_TEST_USER")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("create pool: %v", err)
	}
	return pool
}

func TestMonitorRepository_AddandList(t *testing.T) {
	pool := newPool(t)
	r := NewMonitorRepository(pool)

	ctx, cancelErr := context.WithCancel(context.Background())
	defer cancelErr()

	m, err := monitor.NewMonitor(
		"https://example.com/health",
		5*time.Second,
		30*time.Second)

	if err != nil {
		t.Fatalf("create monitor error: %v", err)
	}

	id, err := r.Add(ctx, m)

	if err != nil {
		t.Fatalf("add monitor: %v", err)
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(), 5*time.Second)

		defer cancel()

		_, err = pool.Exec(cleanupCtx, "DELETE FROM monitors WHERE id=$1", id)
		if err != nil {
			t.Errorf("cleanup monitor %d: %v", id, err)
		}
	}()

	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	ml, err := r.List(ctx)
	if err != nil {
		t.Fatalf("list monitors error: %v", err)
	}
	found := false
	for _, got := range ml {
		if got.ID() != id {
			continue
		}
		found = true

		if got.URL() != m.URL() {
			t.Errorf("URL: want %q, got %q", m.URL(), got.URL())
		}
		if got.Timeout() != m.Timeout() {
			t.Errorf("timeout: want %v, got %v", m.Timeout(), got.Timeout())
		}
		if got.Interval() != m.Interval() {
			t.Errorf("interval: want %v, got %v", m.Interval(), got.Interval())
		}
		break
	}
	if !found {
		t.Fatalf("created monitor %d not found in List", id)
	}
}
