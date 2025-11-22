package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InferRequest struct {
	Prompt string `json:"prompt"`
}

type InferResponse struct {
	ModelRef     string   `json:"modelRef"`
	Prompt       string   `json:"prompt"`
	RouterPod    string   `json:"routerPod"`
	KVEndpoints  []string `json:"kvEndpoints"`
	ProcessingMs int64    `json:"processingMs"`
}

func main() {
	modelRef := getenv("MODEL_REF", "unknown-model")
	maxConcStr := getenv("MAX_CONCURRENCY", "4")
	maxConc, err := strconv.Atoi(maxConcStr)
	if err != nil || maxConc <= 0 {
		log.Printf("invalid MAX_CONCURRENCY=%q, defaulting to 4", maxConcStr)
		maxConc = 4
	}

	kvEndpointsStr := os.Getenv("KV_ENDPOINTS")
	var kvEndpoints []string
	if kvEndpointsStr != "" {
		for _, endp := range strings.Split(kvEndpointsStr, ",") {
			endp = strings.TrimSpace(endp)
			kvEndpoints = append(kvEndpoints, endp)
		}
	}

	log.Printf("starting router with modelRef=%q, maxConcurrency=%d, kvEndpoints=%v",
		modelRef, maxConc, kvEndpoints)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	sem := make(chan struct{}, maxConc)
	var wg sync.WaitGroup

	mux.HandleFunc("/infer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sem <- struct{}{}
		wg.Add(1)
		defer func() {
			<-sem
			wg.Done()
		}()

		start := time.Now()
		var req InferRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		// simulate work
		time.Sleep(50 * time.Millisecond)
		resp := InferResponse{
			ModelRef:     modelRef,
			Prompt:       req.Prompt,
			KVEndpoints:  kvEndpoints,
			ProcessingMs: time.Since(start).Milliseconds(),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})
	addr := ":5678"
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Printf("router listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("router server error: %v", err)
	}

	// Wait for in-flight requests
	wg.Wait()
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
