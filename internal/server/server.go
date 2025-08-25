package server

// genspark: HTTP server with embedded static UI and JSON API

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"genspark-mini/internal/agent"
	"genspark-mini/internal/extract"
)

//go:embed ../../web/*
var webFS embed.FS

// genspark: JSON helpers
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// genspark: Serve starts an HTTP server
func Serve(addr string) error {
	mux := http.NewServeMux()

	// genspark: SPA assets
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// genspark: serve index.html
		b, err := webFS.ReadFile("../../web/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(b)
	})

	mux.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		b, _ := webFS.ReadFile("../../web/app.js")
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		_, _ = w.Write(b)
	})
	mux.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		b, _ := webFS.ReadFile("../../web/styles.css")
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write(b)
	})

	// genspark: health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// genspark: fetch endpoint
	mux.HandleFunc("/api/fetch", func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.Query().Get("url")
		if u == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing url"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		res, err := extract.FetchAndExtract(ctx, u)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	})

	// genspark: summarize endpoint
	mux.HandleFunc("/api/summarize", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		text := string(b)
		if strings.TrimSpace(text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty body"})
			return
		}
		sum := agent.Summarize(text, 5)
		writeJSON(w, http.StatusOK, map[string]any{"summary": sum})
	})

	// genspark: autopilot endpoint - fetch + summarize
	mux.HandleFunc("/api/autopilot", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ URL string `json:"url"` }
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.URL == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing url"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		res, err := extract.FetchAndExtract(ctx, req.URL)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		joined := fmt.Sprintf("%s\n\n%s", res.Title, res.Text)
		sum := agent.Summarize(joined, 5)
		writeJSON(w, http.StatusOK, map[string]any{
			"meta": res,
			"summary": sum,
		})
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           withLogging(mux), // genspark: basic logging
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("[genspark-mini] listening on %s", addr) // genspark
	return srv.ListenAndServe()
}

// genspark: request logger middleware
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("genspark %s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
