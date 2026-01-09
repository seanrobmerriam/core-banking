package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// RequestLogger logs HTTP requests with method, path, status, and duration.
func RequestLogger(log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code
			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(lrw, r)

			// Calculate duration
			duration := time.Since(start)

			// Log request
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", lrw.statusCode).
				Dur("duration", duration).
				Str("user_agent", r.UserAgent()).
				Str("remote_addr", r.RemoteAddr).
				Msg("HTTP request")
		})
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Recover recovers from panics and logs the error.
func Recover(log zerolog.Logger, panicHandler func(ctx context.Context, r *http.Request, recovered interface{})) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					// Log the panic
					log.Error().
						Interface("panic", recovered).
						Str("path", r.URL.Path).
						Str("method", r.Method).
						Msg("Panic recovered")

					// Call custom panic handler if provided
					if panicHandler != nil {
						panicHandler(r.Context(), r, recovered)
					}

					// Return 500 Internal Server Error
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":   "Internal Server Error",
						"message": "A panic occurred while processing your request",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS adds CORS headers to responses.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestIDKey is the context key for request ID.
type RequestIDKey struct{}

// RequestID adds a unique request ID to the context and response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate or extract request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)

		// Add to response headers
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// generateRequestID generates a unique request ID.
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

// JSONContentType sets the Content-Type header to application/json.
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Timeout adds a timeout to the request handler.
func Timeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			var responseStatus int

			// Run the handler in a goroutine
			go func() {
				lrw := &loggingResponseWriter{
					ResponseWriter: w,
					statusCode:     http.StatusOK,
				}
				next.ServeHTTP(lrw, r.WithContext(ctx))
				responseStatus = lrw.statusCode
				close(done)
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Handler completed normally
				w.WriteHeader(responseStatus)
			case <-ctx.Done():
				// Timeout occurred
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Request Timeout",
					"message": "The request took too long to process",
				})
			}
		})
	}
}

// HealthCheckMiddleware adds health check information to requests.
func HealthCheckMiddleware(healthCheck func(r *http.Request) bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add health check info to context
			ctx := context.WithValue(r.Context(), "healthy", healthCheck(r))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestBodyLogger logs the request body for debugging.
func RequestBodyLogger(log zerolog.Logger, maxBodySize int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read and log body for POST/PUT/PATCH requests
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
				if err == nil {
					log.Debug().
						Str("method", r.Method).
						Str("path", r.URL.Path).
						RawJSON("body", body).
						Msg("Request body")
				}
				// Restore the body for the next handler
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MetricsRecorder records metrics for HTTP requests.
type MetricsRecorder struct {
	responseWriter http.ResponseWriter
	body           *bytes.Buffer
	statusCode     int
}

// NewMetricsRecorder creates a new metrics recorder.
func NewMetricsRecorder(w http.ResponseWriter) *MetricsRecorder {
	return &MetricsRecorder{
		responseWriter: w,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

// WriteHeader records the status code.
func (mr *MetricsRecorder) WriteHeader(code int) {
	mr.statusCode = code
	mr.responseWriter.WriteHeader(code)
}

// Write writes data to both the buffer and the original writer.
func (mr *MetricsRecorder) Write(body []byte) (int, error) {
	mr.body.Write(body)
	return mr.responseWriter.Write(body)
}

// GetStatusCode returns the recorded status code.
func (mr *MetricsRecorder) GetStatusCode() int {
	return mr.statusCode
}

// GetBody returns the recorded response body.
func (mr *MetricsRecorder) GetBody() []byte {
	return mr.body.Bytes()
}
