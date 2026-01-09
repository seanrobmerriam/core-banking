package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/core-banking/pkg/config"
	"github.com/core-banking/pkg/logger"
	"github.com/core-banking/pkg/middleware"
)

func main() {
	// Load configuration
	ctx := context.Background()
	cfg, err := config.Load[config.Config](ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.ServiceName, cfg.LogLevel, os.Stdout)
	log := logger.New(cfg.ServiceName)
	log.Info().
		Str("environment", cfg.Environment).
		Int("port", cfg.ServerPort).
		Msg("Starting transaction service (placeholder)")

	// Create router
	router := createRouter(log)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort+2000), // Port 10080
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().
			Str("address", server.Addr).
			Msg("Transaction service server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down transaction service gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Transaction service exited properly")
}

// createRouter creates the HTTP router with all middleware and routes.
func createRouter(log zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Recover(log, nil))
	r.Use(middleware.CORS)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.JSONContentType)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","service":"transaction-service"}`))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Transaction routes (placeholder)
		r.Get("/transactions", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[],"message":"Transaction service - placeholder endpoint"}`))
		})
	})

	return r
}
