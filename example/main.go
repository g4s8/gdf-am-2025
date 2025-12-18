package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "embed"
)

//go:embed index.html
var indexHTML []byte

type ChatRequest struct {
	Model string `json:"model,omitempty"`
	Input string `json:"input"`
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	client := &http.Client{
		Timeout: 0, // streaming: no overall timeout; rely on context cancellation
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	openAI := NewOpenAIClient(client, apiKey)

	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	}))

	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		if req.Input == "" {
			http.Error(w, "input is required", http.StatusBadRequest)
			return
		}
		if req.Model == "" {
			req.Model = "gpt-5.2"
		}

		ctx := r.Context()
		errCh := make(chan error, 1)
		defer close(errCh)

		// Prepare downstream SSE response headers
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("Connection", "keep-alive")
		// Nginx (if any) should not buffer SSE
		w.Header().Set("X-Accel-Buffering", "no")

		stream := openAI.Stream(ctx, req.Input, req.Model, errCh)
		flusher := w.(http.Flusher)
		for part := range stream {
			log.Println("downstream part:", part)
			fmt.Fprint(w, part)
			flusher.Flush()
		}
		select {
		case err := <-errCh:
			if err != nil {
				log.Println("error during streaming:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		default:
		}
		flusher.Flush()
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
