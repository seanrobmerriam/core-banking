package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the application configuration loaded from environment variables.
type Config struct {
	// Server Configuration
	ServerHost string `envconfig:"SERVER_HOST" default:"0.0.0.0"`
	ServerPort int    `envconfig:"SERVER_PORT" default:"8080"`

	// Logging Configuration
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug"`

	// Database Configuration
	DBHost            string        `envconfig:"DB_HOST" required:"true"`
	DBPort            int           `envconfig:"DB_PORT" required:"true"`
	DBUser            string        `envconfig:"DB_USER" required:"true"`
	DBPassword        string        `envconfig:"DB_PASSWORD" required:"true"`
	DBName            string        `envconfig:"DB_NAME" required:"true"`
	DBMaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	DBMaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
	DBConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`

	// NATS Configuration
	NATSURL string `envconfig:"NATS_URL" default:"nats://localhost:4222"`

	// Service Configuration
	ServiceName string `envconfig:"SERVICE_NAME" default:"unknown-service"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
}

// DatabaseConfig returns a DatabaseConfig struct for use with the database package.
func (c *Config) DatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            c.DBHost,
		Port:            c.DBPort,
		User:            c.DBUser,
		Password:        c.DBPassword,
		Name:            c.DBName,
		MaxOpenConns:    c.DBMaxOpenConns,
		MaxIdleConns:    c.DBMaxIdleConns,
		ConnMaxLifetime: c.DBConnMaxLifetime,
	}
}

// DatabaseConfig holds database connection parameters.
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DSN returns the PostgreSQL connection string.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name,
	)
}

// Load loads configuration from environment variables.
// It first attempts to load from a .env file if present, then from system environment variables.
func Load[T Config | DatabaseConfig](ctx context.Context) (T, error) {
	var cfg T

	// Attempt to load from .env file (only in development)
	if os.Getenv("ENVIRONMENT") != "production" {
		if err := godotenv.Load(); err != nil {
			// .env file is optional, don't fail if it doesn't exist
			if !os.IsNotExist(err) {
				return cfg, fmt.Errorf("failed to load .env file: %w", err)
			}
		}
	}

	// Use envconfig to load from environment variables
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to process environment configuration: %w", err)
	}

	// Validate required fields for Config type
	if c, ok := any(&cfg).(Config); ok {
		if err := c.Validate(); err != nil {
			return cfg, fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return cfg, nil
}

// Validate checks that all required configuration fields are set.
func (c *Config) Validate() error {
	if c.DBHost == "" {
		return ErrMissingRequiredConfig("DB_HOST")
	}
	if c.DBPort == 0 {
		return ErrMissingRequiredConfig("DB_PORT")
	}
	if c.DBUser == "" {
		return ErrMissingRequiredConfig("DB_USER")
	}
	if c.DBPassword == "" {
		return ErrMissingRequiredConfig("DB_PASSWORD")
	}
	if c.DBName == "" {
		return ErrMissingRequiredConfig("DB_NAME")
	}
	return nil
}

// ErrMissingRequiredConfig is returned when a required configuration field is missing.
type ErrMissingRequiredConfig string

func (e ErrMissingRequiredConfig) Error() string {
	return fmt.Sprintf("required configuration field is missing: %s", string(e))
}
