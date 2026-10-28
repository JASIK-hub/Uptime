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
type MonitorRepository interface {
	List(ctx context.Context) ([]monitor.Monitor, error)
	Add(ctx context.Context, m monitor.Monitor) (int64, error)
}

type MonitorHandler struct {
	repo MonitorRepository
}

func NewMonitorHandler(r MonitorRepository) *MonitorHandler {
	return &MonitorHandler{
		repo: r,
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
	id, err := h.repo.Add(r.Context(), m)
	if err != nil {
		http.Error(w, "internal database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(id)
}

func (h *MonitorHandler) ListMonitors(w http.ResponseWriter, r *http.Request) {
	monitors, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
