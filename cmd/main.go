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

// Server holds the ichiran manager and readiness state
type Server struct {
	manager *ichiran.Manager
	ready   bool
}

// HandleAnalyze processes text analysis requests
func (s *Server) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	if !s.ready {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	// Extract text from query parameter (e.g., /analyze?text=...)
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Set a timeout for the analysis operation
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Analyze the text using the library
	tokens, err := s.manager.AnalyzeWithContext(ctx, text)
	if err != nil {
		log.Printf("Error analyzing text: %v", err)
		http.Error(w, "Failed to analyze text", http.StatusInternalServerError)
		return
	}

	// Return the tokenized result (customize as needed)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Tokenized: %s\n", tokens.Tokenized())
}

// HandleHealth provides a readiness check
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if s.ready {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(w, "Service Unavailable")
	}
}

func main() {
	server := &Server{}

	// Register endpoints
	http.HandleFunc("/health", server.HandleHealth)
	http.HandleFunc("/analyze", server.HandleAnalyze)

	// Start the HTTP server
	srv := &http.Server{Addr: ":8080"}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Initialize the ichiran library in a separate goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// Initialize with context and custom project name
		manager, err := ichiran.NewManager(ctx, ichiran.WithProjectName("go-ichiran-api"))
		if err != nil {
			log.Fatalf("Failed to initialize ichiran: %v", err)
		}
		server.manager = manager
		server.ready = true
		log.Println("Ichiran initialized successfully")
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

	// Clean up the ichiran manager
	if server.manager != nil {
		server.manager.Close()
	}
	log.Println("Server exited")
}