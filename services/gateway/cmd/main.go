package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	const GATEWAY_PORT = "7001"
	mux := http.NewServeMux()
	mux.HandleFunc("/health", gatewatHealthHandler)
	mux.HandleFunc("/logs/", logsProxyHandler)
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

func logsProxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Proxying request to logs-service")
	target, err := url.Parse("http://localhost:7017")
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	// Override Director to rewrite path
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Remove "/logs" prefix
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/logs")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
	}
	proxy.ServeHTTP(w, r)
}
