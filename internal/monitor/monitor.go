package monitor

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"
)

type Monitor struct {
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

type MonitorStore struct {
	Monitors map[int]Monitor
	NextId   int
	mu       sync.Mutex
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

func NewMonitorStore(m Monitor) *MonitorStore {
	return &MonitorStore{
		NextId:   1,
		Monitors: make(map[int]Monitor),
	}
}

func (s *MonitorStore) Add(m Monitor) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.NextId
	s.Monitors[id] = m
	s.NextId++
	return id
}
