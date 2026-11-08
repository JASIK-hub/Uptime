-- +goose Up
ALTER TABLE monitors_history
DROP CONSTRAINT monitors_history_duration_ms_check,
ALTER COLUMN duration_ms SET NOT NULL,
ADD CONSTRAINT monitors_history_duration_ms_check
    CHECK (duration_ms>=0);

-- +goose Down
ALTER TABLE monitors_history
DROP CONSTRAINT monitors_history_duration_ms_check,
ALTER COLUMN duration_ms DROP NOT NULL,
ADD CONSTRAINT monitors_history_duration_ms_check
        CHECK (duration_ms > 0);
