package monitor

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

type Monitor struct {
	url     url.URL
	timeout time.Duration
}

type CheckResult struct {
	Success    bool
	StartTime  time.Time
	StatusCode *int
	Duration   time.Duration
}

var (
	ErrInvalidURL     = errors.New("invalid URL Format")
	ErrInvalidTimeout = errors.New("invalid timeout value")
)

func NewMonitor(rawURL string, timeout time.Duration) (Monitor, error) {
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
	return Monitor{url: *u, timeout: timeout}, nil
}
