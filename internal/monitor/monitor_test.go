package monitor

import (
	"errors"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestNewMonitor(t *testing.T) {
	tt := map[string]struct {
		url      string
		timeout  time.Duration
		interval time.Duration
		want     Monitor
		err      error
	}{
		"complete URL": {
			url:      "https://example.com",
			timeout:  5 * time.Second,
			interval: 60 * time.Second,
			want:     Monitor{url: url.URL{Scheme: "https", Host: "example.com"}, timeout: 5 * time.Second, interval: 60 * time.Second},
			err:      nil,
		},
		"complete URL with path": {
			url:      "https://example.com/health",
			timeout:  time.Second,
			interval: 60 * time.Second,
			want:     Monitor{url: url.URL{Scheme: "https", Host: "example.com", Path: "/health"}, timeout: time.Second, interval: 60 * time.Second},
			err:      nil,
		},
		"URL without scheme": {
			url:      "example.com",
			timeout:  time.Second,
			interval: 60 * time.Second,
			want:     Monitor{},
			err:      ErrInvalidURL,
		},
		"URL with invalid scheme": {
			url:      "ftp://example.com",
			timeout:  time.Second,
			interval: 60 * time.Second,
			want:     Monitor{},
			err:      ErrInvalidURL,
		},
		"URL without host": {
			url:      "https://",
			timeout:  time.Second,
			interval: 60 * time.Second,
			want:     Monitor{},
			err:      ErrInvalidURL,
		},
		"invalid timeout value": {
			url:      "https://example.com",
			timeout:  0,
			interval: 60 * time.Second,
			want:     Monitor{},
			err:      ErrInvalidTimeout,
		},
		"invalid interval value": {
			url:      "https://example.com",
			timeout:  5 * time.Second,
			interval: 0,
			want:     Monitor{},
			err:      ErrInvalidInterval,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got, err := NewMonitor(tc.url, tc.timeout, tc.interval)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected %v, got %v", tc.err, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %+v, got %+v", tc.want, got)
			}
		})
	}
}
