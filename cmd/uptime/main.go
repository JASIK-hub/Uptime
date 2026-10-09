package main

import (
	"log"
	"net/http"
	"uptime/internal/api"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /monitor", api.CreateMonitor)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
