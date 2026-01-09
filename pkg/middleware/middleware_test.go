package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// testLogger returns a logger that writes to a buffer for testing.
func testLogger() (zerolog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	log := zerolog.New(buf).With().Timestamp().Logger()
	return log, buf
}

func newTestRouter(log zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(RequestLogger(log))
	r.Use(Recover(log, nil))
	return r
}

func TestRequestLogger(t *testing.T) {
	log, _ := testLogger()
	router := newTestRouter(log)

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestLoggerWithPanic(t *testing.T) {
	log, _ := testLogger()
	recoveryCalled := false

	router := chi.NewRouter()
	router.Use(Recover(log, func(ctx context.Context, r *http.Request, recovered interface{}) {
		recoveryCalled = true
	}))
	router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		router.ServeHTTP(w, req)
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, recoveryCalled)
}

func TestRecover(t *testing.T) {
	log, _ := testLogger()

	t.Run("recovers from panic", func(t *testing.T) {
		router := chi.NewRouter()
		router.Use(Recover(log, nil))
		router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
			panic("panic!")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()

		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("custom panic handler is called", func(t *testing.T) {
		handlerCalled := false
		panicHandler := func(ctx context.Context, r *http.Request, recovered interface{}) {
			handlerCalled = true
		}

		router := chi.NewRouter()
		router.Use(Recover(log, panicHandler))
		router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
			panic("panic!")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.True(t, handlerCalled)
	})

	t.Run("no panic handler", func(t *testing.T) {
		router := chi.NewRouter()
		router.Use(Recover(log, nil))
		router.Get("/ok", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCORS(t *testing.T) {
	router := chi.NewRouter()
	router.Use(CORS)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSHeaders(t *testing.T) {
	router := chi.NewRouter()
	router.Use(CORS)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Accept")
}

func TestRequestID(t *testing.T) {
	router := chi.NewRouter()
	router.Use(RequestID)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		w.Header().Set("X-Request-ID", requestID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestJSONContentType(t *testing.T) {
	router := chi.NewRouter()
	router.Use(JSONContentType)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func newTestLoggerMiddlewareRouter(log zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(RequestLogger(log))
	return r
}

func TestLoggerMiddlewareWithResponseBody(t *testing.T) {
	log, _ := testLogger()
	router := newTestLoggerMiddlewareRouter(log)

	router.Get("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTimeout(t *testing.T) {
	router := chi.NewRouter()
	router.Use(Timeout(5 * time.Second))
	router.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTimeoutExceeds(t *testing.T) {
	router := chi.NewRouter()
	router.Use(Timeout(50 * time.Millisecond))
	router.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestMultipleMiddlewares(t *testing.T) {
	log, _ := testLogger()
	callOrder := []string{}

	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder = append(callOrder, "first")
			next.ServeHTTP(w, r)
		})
	})
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder = append(callOrder, "second")
			next.ServeHTTP(w, r)
		})
	})
	router.Use(RequestLogger(log))
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		callOrder = append(callOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"first", "second", "handler"}, callOrder)
}

func TestHealthCheckMiddleware(t *testing.T) {
	router := chi.NewRouter()
	router.Use(HealthCheckMiddleware(func(r *http.Request) bool {
		return true
	}))
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
