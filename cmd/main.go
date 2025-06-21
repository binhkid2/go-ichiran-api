package main

import (
	"context"
	"fmt"
	"github.com/binhkid2/go-ichiran-api/internal/config"
	"github.com/binhkid2/go-ichiran-api/internal/handler"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize logger
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Ichiran
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := ichiran.InitWithContext(ctx); err != nil {
		log.Fatalf("Failed to initialize Ichiran: %v", err)
	}
	defer ichiran.Close()

	// Initialize router and handlers
	router := mux.NewRouter()
	textHandler := handler.NewTextHandler(log)
	router.HandleFunc("/api/analyze", textHandler.Analyze).Methods("POST")

	// Create server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Starting server on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Create shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Info("Server stopped")
}
