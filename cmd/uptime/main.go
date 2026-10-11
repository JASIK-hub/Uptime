package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"uptime/internal/api"
	"uptime/internal/monitor"
	"uptime/internal/scheduler"
)

func main() {
	store := monitor.NewMonitorStore()
	handler := api.NewMonitorHandler(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go scheduler.RunScheduler(ctx, store)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /monitor", handler.CreateMonitor)
	mux.HandleFunc("GET /monitors", handler.ListMonitors)
	fmt.Print("Server started")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
