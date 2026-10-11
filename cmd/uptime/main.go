package main

import (
	"log"
	"net/http"
	"uptime/internal/api"
	"uptime/internal/monitor"
)

func main() {
	store := monitor.NewMonitorStore()
	handler := api.NewMonitorHandler(store)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /monitor", handler.CreateMonitor)
	mux.HandleFunc("GET /monitors", handler.ListMonitors)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
