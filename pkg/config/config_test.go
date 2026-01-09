package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfigDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "standard connection",
			config: DatabaseConfig{
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
			config: DatabaseConfig{
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

func TestConfigDatabaseConfig(t *testing.T) {
	cfg := Config{
		DBHost:            "db.example.com",
		DBPort:            5433,
		DBUser:            "admin",
		DBPassword:        "secret",
		DBName:            "production",
		DBMaxOpenConns:    100,
		DBMaxIdleConns:    10,
		DBConnMaxLifetime: 0,
	}

	dbCfg := cfg.DatabaseConfig()

	assert.Equal(t, "db.example.com", dbCfg.Host)
	assert.Equal(t, 5433, dbCfg.Port)
	assert.Equal(t, "admin", dbCfg.User)
	assert.Equal(t, "secret", dbCfg.Password)
	assert.Equal(t, "production", dbCfg.Name)
	assert.Equal(t, 100, dbCfg.MaxOpenConns)
	assert.Equal(t, 10, dbCfg.MaxIdleConns)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "user",
				DBPassword: "pass",
				DBName:     "db",
			},
			wantErr: false,
		},
		{
			name: "missing DB_HOST",
			cfg: Config{
				DBPort:     5432,
				DBUser:     "user",
				DBPassword: "pass",
				DBName:     "db",
			},
			wantErr: true,
			errMsg:  "required configuration field is missing: DB_HOST",
		},
		{
			name: "missing DB_PORT",
			cfg: Config{
				DBHost:     "localhost",
				DBUser:     "user",
				DBPassword: "pass",
				DBName:     "db",
			},
			wantErr: true,
			errMsg:  "required configuration field is missing: DB_PORT",
		},
		{
			name: "missing DB_USER",
			cfg: Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBPassword: "pass",
				DBName:     "db",
			},
			wantErr: true,
			errMsg:  "required configuration field is missing: DB_USER",
		},
		{
			name: "missing DB_PASSWORD",
			cfg: Config{
				DBHost: "localhost",
				DBPort: 5432,
				DBUser: "user",
				DBName: "db",
			},
			wantErr: true,
			errMsg:  "required configuration field is missing: DB_PASSWORD",
		},
		{
			name: "missing DB_NAME",
			cfg: Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "user",
				DBPassword: "pass",
			},
			wantErr: true,
			errMsg:  "required configuration field is missing: DB_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
