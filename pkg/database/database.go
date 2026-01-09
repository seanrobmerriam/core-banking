package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"

	"github.com/core-banking/pkg/config"
	"github.com/core-banking/pkg/logger"
)

// DB wraps the sql.DB with additional functionality.
type DB struct {
	*sql.DB
	cfg config.DatabaseConfig
}

// NewDatabase creates a new database connection pool.
func NewDatabase(ctx context.Context, cfg config.DatabaseConfig, log *zerolog.Logger) (*DB, error) {
	dsn := cfg.DSN()
	log.Info().Msg("Connecting to PostgreSQL database...")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection with context timeout
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(connectCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Int("max_open_conns", cfg.MaxOpenConns).
		Int("max_idle_conns", cfg.MaxIdleConns).
		Dur("conn_max_lifetime", cfg.ConnMaxLifetime).
		Msg("Database connection established successfully")

	return &DB{DB: db, cfg: cfg}, nil
}

// HealthCheck performs a health check on the database connection.
func (db *DB) HealthCheck(ctx context.Context) error {
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Status returns the current status of the database connection.
func (db *DB) Status(ctx context.Context) (HealthStatus, error) {
	stats := db.Stats()

	if err := db.PingContext(ctx); err != nil {
		return HealthStatus{
			Status:           "unhealthy",
			Database:         "disconnected",
			OpenConnections:  int(stats.OpenConnections),
			InUse:            int(stats.InUse),
			Idle:             int(stats.Idle),
			WaitCount:        int(stats.WaitCount),
			WaitDuration:     stats.WaitDuration,
			MaxOpenConnections: int(stats.MaxOpenConnections),
		}, nil
	}

	return HealthStatus{
		Status:             "healthy",
		Database:           "connected",
		OpenConnections:    int(stats.OpenConnections),
		InUse:              int(stats.InUse),
		Idle:               int(stats.Idle),
		WaitCount:          int(stats.WaitCount),
		WaitDuration:       stats.WaitDuration,
		MaxOpenConnections: int(stats.MaxOpenConnections),
	}, nil
}

// HealthStatus represents the health status of the database.
type HealthStatus struct {
	Status             string        `json:"status"`
	Database           string        `json:"database"`
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int           `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxOpenConnections int           `json:"max_open_connections"`
}

// Close closes the database connection and logs the closure.
func (db *DB) Close() error {
	log := logger.Global
	log.Info().
		Str("host", db.cfg.Host).
		Str("database", db.cfg.Name).
		Msg("Closing database connection")

	if err := db.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
