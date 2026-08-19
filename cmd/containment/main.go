package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

func main() {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "semiconductor-excursion-containment"
	}
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(healthResponse{
			Status:  "ok",
			Service: serviceName,
			Time:    time.Now().UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("GET /api/v1/summary", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"capabilities": []string{"containment", "sampling", "review", "investigation"},
			"ready":        true,
		})
	})

	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("containment service listening on %s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
