package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/core-banking/pkg/config"
)

func TestDatabaseConfigDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   config.DatabaseConfig
		expected string
	}{
		{
			name: "standard connection",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Name:     "testdb",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=testdb sslmode=disable",
		},
		{
			name: "custom port",
			config: config.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "secret",
				Name:     "production",
			},
			expected: "host=db.example.com port=5433 user=admin password=secret dbname=production sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.DSN())
		})
	}
}

func TestHealthStatus(t *testing.T) {
	t.Run("health status fields", func(t *testing.T) {
		status := HealthStatus{
			Status:             "healthy",
			Database:           "connected",
			OpenConnections:    10,
			InUse:              3,
			Idle:               7,
			WaitCount:          0,
			WaitDuration:       0,
			MaxOpenConnections: 25,
		}

		assert.Equal(t, "healthy", status.Status)
		assert.Equal(t, "connected", status.Database)
		assert.Equal(t, 10, status.OpenConnections)
		assert.Equal(t, 3, status.InUse)
		assert.Equal(t, 7, status.Idle)
		assert.Equal(t, 25, status.MaxOpenConnections)
	})

	t.Run("unhealthy status", func(t *testing.T) {
		status := HealthStatus{
			Status:             "unhealthy",
			Database:           "disconnected",
			OpenConnections:    0,
			InUse:              0,
			Idle:               0,
			WaitCount:          0,
			WaitDuration:       0,
			MaxOpenConnections: 0,
		}

		assert.Equal(t, "unhealthy", status.Status)
		assert.Equal(t, "disconnected", status.Database)
	})
}

func TestConnectionPoolConfiguration(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "user",
		Password:        "pass",
		Name:            "db",
		MaxOpenConns:    100,
		MaxIdleConns:    20,
		ConnMaxLifetime: 0,
	}

	assert.Equal(t, 100, cfg.MaxOpenConns)
	assert.Equal(t, 20, cfg.MaxIdleConns)
}
