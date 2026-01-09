package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	// Global is the default logger instance used throughout the application.
	Global zerolog.Logger

	// serviceName is used to set the service name in log output.
	serviceName = "unknown-service"
)

func init() {
	// Initialize global logger with defaults
	Global = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}

// Init initializes the global logger with the given service name and configuration.
func Init(name string, level string, output io.Writer) {
	serviceName = name

	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Create new logger
	Global = zerolog.New(output).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger().
		Level(logLevel)
}

// New creates a new logger with the service name already set.
func New(name string) zerolog.Logger {
	return Global.With().Str("service", name).Logger()
}

// WithService creates a logger with the specified service name.
func WithService(name string) zerolog.Logger {
	return Global.With().Str("service", name).Logger()
}

// RequestIDKey is the context key for the request ID.
type RequestIDKey struct{}

// WithRequestID generates a new request ID and adds it to the context.
func WithRequestID(ctx context.Context) (context.Context, string) {
	requestID := uuid.New().String()
	newCtx := context.WithValue(ctx, RequestIDKey{}, requestID)
	return newCtx, requestID
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// Ctx returns a logger event that includes the request ID from context.
func Ctx(ctx context.Context) *zerolog.Event {
	requestID := GetRequestID(ctx)
	event := Global.Info()
	if requestID != "" {
		event = event.Str("request_id", requestID)
	}
	return event
}

// WithTrace adds caller information for trace purposes.
// Uses runtime information to get the file and line number.
func WithTrace() *zerolog.Event {
	_, file, line, _ := runtime.Caller(1)
	// Get just the relative path
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}
	return Global.Trace().Caller().Str("file", fmt.Sprintf("%s:%d", file, line))
}

// LogRequest logs an HTTP request with method, path, status, and duration.
func LogRequest(log zerolog.Logger, method, path string, status int, duration time.Duration) {
	log.Info().
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Dur("duration", duration).
		Msg("HTTP request processed")
}

// LogError logs an error with optional context.
func LogError(log zerolog.Logger, err error, msg string) {
	log.Error().Err(err).Msg(msg)
}

// DevelopmentConfig returns a zerolog configuration for development.
func DevelopmentConfig() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s |", i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf(">> %s", i)
		},
		FormatCaller: func(i interface{}) string {
			return fmt.Sprintf("[%s]", i)
		},
	}
}

// ProductionConfig returns a zerolog configuration for production (JSON output).
func ProductionConfig(output io.Writer) zerolog.Logger {
	return zerolog.New(output).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}

// SetGlobalLogger sets the global logger instance.
func SetGlobalLogger(log zerolog.Logger) {
	Global = log
}

// GetGlobalLogger returns the global logger instance.
func GetGlobalLogger() zerolog.Logger {
	return Global
}
