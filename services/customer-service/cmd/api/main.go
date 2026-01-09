package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/core-banking/pkg/config"
	"github.com/core-banking/pkg/database"
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
		Msg("Starting customer service")

	// Initialize database
	Global.DB, err = database.NewDatabase(ctx, cfg.DatabaseConfig(), &log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer Global.DB.Close()

	// Verify database health
	if err := Global.DB.HealthCheck(ctx); err != nil {
		log.Fatal().Err(err).Msg("Database health check failed")
	}
	log.Info().Msg("Database health check passed")

	// Create router
	router := createRouter(log)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().
			Str("address", server.Addr).
			Msg("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}

// createRouter creates the HTTP router with all middleware and routes.
func createRouter(log zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Recover(log, nil))
	r.Use(middleware.CORS)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.JSONContentType)

	// Health check endpoint (no authentication required)
	r.Get("/health", healthHandler(log))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Customer routes
		r.Route("/customers", func(r chi.Router) {
			r.Get("/", listCustomersHandler)
			r.Post("/", createCustomerHandler)
			r.Get("/{id}", getCustomerHandler)
			r.Put("/{id}", updateCustomerHandler)
			r.Delete("/{id}", deleteCustomerHandler)
		})
	})

	return r
}

// healthHandler returns the health status of the service.
func healthHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check database connectivity
		status, err := Global.DB.Status(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Database health check failed")
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}

		// Return health status
		statusResponse := map[string]interface{}{
			"status":    "healthy",
			"service":   "customer-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"database": map[string]interface{}{
				"status":           status.Status,
				"open_connections": status.OpenConnections,
				"idle_connections": status.Idle,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(statusResponse); err != nil {
			log.Error().Err(err).Msg("Failed to encode health response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// listCustomersHandler returns a list of customers.
func listCustomersHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual database query
	customers := []map[string]interface{}{
		{
			"id":         "1",
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john.doe@example.com",
		},
		{
			"id":         "2",
			"first_name": "Jane",
			"last_name":  "Smith",
			"email":      "jane.smith@example.com",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      customers,
		"total":     len(customers),
		"page":      1,
		"page_size": 10,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// createCustomerHandler creates a new customer.
func createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var req createCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		http.Error(w, "Missing required fields: first_name, last_name, email", http.StatusBadRequest)
		return
	}

	// TODO: Insert into database
	customer := map[string]interface{}{
		"id":         generateID(),
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"email":      req.Email,
		"phone":      req.Phone,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(customer); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getCustomerHandler returns a customer by ID.
func getCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	// TODO: Fetch from database
	customer := map[string]interface{}{
		"id":         id,
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john.doe@example.com",
		"phone":      "+1234567890",
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(customer); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// updateCustomerHandler updates an existing customer.
func updateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	var req updateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Update in database
	customer := map[string]interface{}{
		"id":         id,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"email":      req.Email,
		"phone":      req.Phone,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(customer); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// deleteCustomerHandler deletes a customer by ID.
func deleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	// TODO: Delete from database
	w.WriteHeader(http.StatusNoContent)
}

// Request/Response types

type createCustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
}

type updateCustomerRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// Helper functions

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Global database instance for health check
var Global = struct {
	DB *database.DB
}{}
