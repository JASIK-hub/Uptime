-- +goose Up
CREATE TABLE monitors(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    url TEXT NOT NULL,
    timeout_seconds INTEGER NOT NULL CHECK (timeout_seconds>0),
    interval_seconds INTEGER NOT NULL CHECK (interval_seconds>0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE MONITORS;
