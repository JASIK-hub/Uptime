package repository

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	db "uptime/internal/infra/postgres/sqlc"
	"uptime/internal/monitor"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *MonitorRepository) SaveCheck(
	ctx context.Context,
	monitorId int64,
	result monitor.CheckResult,
	errMsg error,
) error {
	ms := result.Duration.Milliseconds()
	code := *result.StatusCode
	if ms < 0 || ms > math.MaxInt32 {
		return fmt.Errorf("check duration out of range: %d ms", ms)
	}

	if code < 100 || code > 599 {
		return fmt.Errorf("invalid HTTP status code: %d", code)
	}

	params := db.AddMonitorCheckParams{
		MonitorID: monitorId,
		DurationMs: pgtype.Int4{
			Int32: int32(ms),
			Valid: true,
		},
		StartedAt: pgtype.Timestamptz{
			Time:  result.StartTime,
			Valid: true,
		},
		Success: result.Success,
	}
	if result.StatusCode != nil {
		params.StatusCode = pgtype.Int4{
			Int32: int32(code),
			Valid: true,
		}
	}
	if errMsg != nil {
		params.ErrorMessage = pgtype.Text{
			String: errMsg.Error(),
			Valid:  true,
		}
	}

	err := r.queries.AddMonitorCheck(ctx, params)
	if err != nil {
		slog.Warn("could not add monitor to history", "error", err)
		return err
	}
	return nil
}
