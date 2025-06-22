package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
)

var ichiranReady atomic.Bool

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic: %v", r)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if !ichiranReady.Load() {
		http.Error(w, "Service not ready", http.StatusServiceUnavailable)
		return
	}

	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	tokens, err := ichiran.Analyze(text)
	if err != nil {
		log.Printf("Error analyzing text %q: %v", text, err)
		http.Error(w, "Failed to analyze text", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Tokenized: %s\n", tokens.Tokenized())
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if !ichiranReady.Load() {
		http.Error(w, "Ichiran not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ichiran.AnalyzeWithContext(ctx, "テスト")
	if err != nil {
		log.Printf("Health check failed: %v", err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Ichiran server is running.")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/analyze", handleAnalyze)
	mux.HandleFunc("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      recoverMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("Server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	go func() {
		log.Println("Initializing ichiran with recreate...")
		for i := 0; i < 3; i++ {
			if err := ichiran.InitRecreate(true); err != nil {
				log.Printf("Failed to initialize ichiran (attempt %d): %v", i+1, err)
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}
			log.Println("Ichiran initialized successfully")
			ichiranReady.Store(true)
			return
		}
		log.Fatal("Failed to initialize ichiran after retries")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	ichiran.Close()
	log.Println("Server exited cleanly")
}
