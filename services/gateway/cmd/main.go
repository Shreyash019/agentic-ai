package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	const GATEWAY_PORT = "7001"
	mux := http.NewServeMux()
	mux.HandleFunc("/health", gatewatHealthHandler)
	log.Println("Gateway running on ", GATEWAY_PORT)
	err := http.ListenAndServe(":"+GATEWAY_PORT, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func gatewatHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "Gateway OK",
	}
	log.Println("Gateway Health is OK")
	json.NewEncoder(w).Encode(response)
}
