package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// generateID returns a random UUID v4 string.
// Used to stamp X-Request-ID when the upstream client did not supply one.
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely; fall back to a timestamp-based hex string.
		return fmt.Sprintf("%x", b)
	}
	// Set UUID v4 version and variant bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// requestIDMiddleware stamps every inbound request with X-Request-ID and
// X-Trace-ID before the request is forwarded to a downstream service.
//
// Header semantics:
//
//	X-Request-ID — unique per HTTP request; echoed back to the client.
//	               A client may supply its own value; we honour it.
//	X-Trace-ID   — distributed trace token that spans multiple services.
//	               Today it equals X-Request-ID. When nginx becomes the LB it
//	               will inject X-Trace-ID itself; the gateway will simply
//	               forward whatever value arrives.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateID()
		}

		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			// No distributed trace context yet — seed it from the request ID.
			traceID = requestID
		}

		// Propagate to downstream service.
		r.Header.Set("X-Request-ID", requestID)
		r.Header.Set("X-Trace-ID", traceID)

		// Echo back to the caller so clients can correlate their requests.
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

func main() {
	const GATEWAY_PORT = "7001"
	mux := http.NewServeMux()
	mux.HandleFunc("/health", gatewatHealthHandler)
	mux.HandleFunc("/logs/", logsProxyHandler)
	mux.HandleFunc("/auth/", authServiceProxyHandler)
	log.Println("Gateway running on ", GATEWAY_PORT)
	err := http.ListenAndServe(":"+GATEWAY_PORT, requestIDMiddleware(mux))
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

func authServiceProxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Proxy Request to Auth Service")
	target, err := url.Parse("http://localhost:7009")
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/auth")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}

	}
	proxy.ServeHTTP(w, r)
}
