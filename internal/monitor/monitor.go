package monitor

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

type Monitor struct {
	id       int64
	url      url.URL
	timeout  time.Duration
	interval time.Duration
}

type CheckResult struct {
	Success    bool
	StartTime  time.Time
	StatusCode *int
	Duration   time.Duration
}

var (
	ErrInvalidURL      = errors.New("invalid URL format")
	ErrInvalidTimeout  = errors.New("invalid timeout value")
	ErrInvalidInterval = errors.New("monitor interval must be positive")
)

func NewMonitor(rawURL string, timeout time.Duration, interval time.Duration) (Monitor, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return Monitor{}, fmt.Errorf("parse monitor URL: %w", err)
	}
	if u.Hostname() == "" {
		return Monitor{}, ErrInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return Monitor{}, ErrInvalidURL
	}
	if timeout <= 0 {
		return Monitor{}, ErrInvalidTimeout
	}
	if interval <= 0 {
		return Monitor{}, ErrInvalidInterval
	}
	return Monitor{url: *u, timeout: timeout, interval: interval}, nil
}

func (m Monitor) URL() string {
	return m.url.String()
}

func (m Monitor) Interval() time.Duration {
	return m.interval
}

func (m Monitor) Timeout() time.Duration {
	return m.timeout
}

func (m Monitor) ID() int64 {
	return m.id
}

func RestoreMonitor(id int64, rawUrl string, timeout, interval time.Duration) (Monitor, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return Monitor{}, fmt.Errorf("restore monitor parse url: %w", err)
	}
	return Monitor{
		id:       id,
		url:      *u,
		timeout:  timeout,
		interval: interval,
	}, nil
}
