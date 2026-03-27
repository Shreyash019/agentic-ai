package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	const LOGGER_SERVICE_PORT = "7017"
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	log.Println("Service Running on Port", LOGGER_SERVICE_PORT)
	err := http.ListenAndServe(":"+LOGGER_SERVICE_PORT, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Service health is OK")
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "OK",
	}
	json.NewEncoder(w).Encode(response)
}
