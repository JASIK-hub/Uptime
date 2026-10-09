package api

import (
	"encoding/json"
	"net/http"
)

type createMonitorRequest struct {
	URL             string `json:"url"`
	IntervalSeconds int    `json:"intervalSeconds"`
	TimeoutSeconds  int    `json:"timeoutSeconds"`
}

func CreateMonitor(w http.ResponseWriter, r *http.Request) {
	var input createMonitorRequest

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(input)
}
