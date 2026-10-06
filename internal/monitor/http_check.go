package monitor

import (
	"net/http"
	"time"
)

func HTTPCheck(m Monitor) (CheckResult, error) {
	url := m.url.String()
	client := http.Client{Timeout: m.timeout}
	startTime := time.Now()
	resp, err := client.Get(url)
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
