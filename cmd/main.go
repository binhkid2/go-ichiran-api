package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
)

func main() {
	// Initialize ichiran in a separate goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := ichiran.InitWithContext(ctx); err != nil {
			log.Fatalf("Failed to initialize ichiran: %v", err)
		}
		log.Println("Ichiran initialized successfully")
	}()

	// Register HTTP handlers
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/analyze", handleAnalyze)

	// Start the HTTP server
	srv := &http.Server{Addr: ":8080"}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown the HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	// Clean up ichiran
	ichiran.Close()
	log.Println("Server exited")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	// Basic health check (could be enhanced to verify ichiran readiness)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	// Extract text from query parameter (e.g., /analyze?text=...)
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Set a timeout for the analysis operation
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Analyze the text using ichiran
	tokens, err := ichiran.AnalyzeWithContext(ctx, text)
	if err != nil {
		log.Printf("Error analyzing text: %v", err)
		http.Error(w, "Failed to analyze text", http.StatusInternalServerError)
		return
	}

	// Return tokenized result (mimicking example output)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Tokenized: %s\n", tokens.Tokenized())
}
