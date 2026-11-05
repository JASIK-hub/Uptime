-- +goose Up
CREATE TABLE monitors_history(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    monitor_id BIGINT NOT NULL,
    duration_ms INTEGER CHECK (duration_ms>0),
    started_at TIMESTAMPTZ NOT NULL,
    success BOOLEAN NOT NULL,
    status_code INTEGER CHECK (status_code BETWEEN 100 and 599),
    error_message TEXT,
    CONSTRAINT fk_monitors_history_monitors FOREIGN KEY (monitor_id) REFERENCES monitors(id)
);

CREATE INDEX idx_monitors_history_monitor ON monitors_history(monitor_id,started_at DESC,id DESC);

-- +goose Down
DROP TABLE monitors_history;
