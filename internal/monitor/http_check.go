package monitor

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func HTTPCheck(ctx context.Context, m Monitor) (CheckResult, error) {
	url := m.url.String()
	client := http.Client{Timeout: m.timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return CheckResult{}, fmt.Errorf("create HTTP request: %w", err)
	}
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return CheckResult{
			Success:    false,
			StartTime:  startTime,
			Duration:   duration,
			StatusCode: nil,
		}, err
	}
	defer resp.Body.Close()

	result := CheckResult{
		Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		StartTime:  startTime,
		Duration:   duration,
		StatusCode: &resp.StatusCode,
	}
	return result, nil
}
