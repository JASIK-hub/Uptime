package api

import (
	"encoding/json"
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
