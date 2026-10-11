package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"uptime/internal/monitor"
)

type createMonitorRequest struct {
	URL             string `json:"url"`
	IntervalSeconds int    `json:"intervalSeconds"`
	TimeoutSeconds  int    `json:"timeoutSeconds"`
}

type monitorResponse struct {
	ID              int    `json:"id"`
	URL             string `json:"url"`
	IntervalSeconds int64  `json:"intervalSeconds"`
	TimeoutSeconds  int64  `json:"timeoutSeconds"`
}

type MonitorHandler struct {
	store *monitor.MonitorStore
}

func NewMonitorHandler(s *monitor.MonitorStore) *MonitorHandler {
	return &MonitorHandler{
		store: s,
	}
}

func (h *MonitorHandler) CreateMonitor(w http.ResponseWriter, r *http.Request) {
	var input createMonitorRequest

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	m, err := monitor.NewMonitor(input.URL, time.Duration(input.TimeoutSeconds)*time.Second, time.Duration(input.IntervalSeconds)*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := h.store.Add(m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(id)
}

func (h *MonitorHandler) ListMonitors(w http.ResponseWriter, r *http.Request) {
	monitors := h.store.List()
	serialized := []monitorResponse{}
	for id, m := range monitors {
		monitor := monitorResponse{
			ID:              id,
			URL:             m.URL(),
			TimeoutSeconds:  int64(m.Timeout() / time.Second),
			IntervalSeconds: int64(m.Interval() / time.Second),
		}
		serialized = append(serialized, monitor)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serialized)
}

func RunMonitor(ctx context.Context, m monitor.Monitor) error {
	check := func() {
		c, err := monitor.HTTPCheck(ctx, m)
		if err != nil {
			log.Printf("check %s failed: %v", m.URL(), err)
			return
		}
		log.Printf("check %s: success=%t, duration=%v",
			m.URL(), c.Success, c.Duration)
	}
	ticker := time.NewTicker(m.Interval())
	defer ticker.Stop()

	check()

	for {
		select {
		case <-ticker.C:
			if ctx.Err() != nil {
				return ctx.Err()
			}
			check()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
