package logger

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	Init("test-service", "debug", DevNull())
	assert.Equal(t, "test-service", serviceName)
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	newCtx, requestID := WithRequestID(ctx)

	assert.NotEmpty(t, requestID)
	assert.Equal(t, requestID, GetRequestID(newCtx))
	assert.Empty(t, GetRequestID(ctx))
}

func TestGetRequestID(t *testing.T) {
	t.Run("with request ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), RequestIDKey{}, "test-request-id")
		assert.Equal(t, "test-request-id", GetRequestID(ctx))
	})

	t.Run("without request ID", func(t *testing.T) {
		ctx := context.Background()
		assert.Empty(t, GetRequestID(ctx))
	})

	t.Run("with wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), RequestIDKey{}, 12345)
		assert.Empty(t, GetRequestID(ctx))
	})
}

func TestCtx(t *testing.T) {
	t.Run("with request ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), RequestIDKey{}, "test-123")
		event := Ctx(ctx)
		assert.NotNil(t, event)
	})

	t.Run("without request ID", func(t *testing.T) {
		ctx := context.Background()
		event := Ctx(ctx)
		assert.NotNil(t, event)
	})
}

func TestNew(t *testing.T) {
	Init("main-service", "debug", DevNull())
	log := New("customer-service")
	assert.NotNil(t, log)
}

func TestWithService(t *testing.T) {
	Init("main-service", "debug", DevNull())
	log := WithService("account-service")
	assert.NotNil(t, log)
}

func TestServiceName(t *testing.T) {
	Init("my-service", "info", DevNull())
	assert.Equal(t, "my-service", serviceName)
}

func TestDevNull(t *testing.T) {
	writer := DevNull()
	assert.NotNil(t, writer)
}

func DevNull() io.Writer {
	return &nullWriter{}
}

type nullWriter struct{}

func (w *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestRequestIDKeyUniqueness(t *testing.T) {
	key1 := RequestIDKey{}
	key2 := RequestIDKey{}
	assert.Equal(t, key1, key2)
}

func TestLoggerInitializationModes(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
		{"invalid level defaults to info", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init("test", tt.level, DevNull())
			// Just verify no panic occurs
			_ = New("test")
		})
	}
}

func TestWithTrace(t *testing.T) {
	Init("trace-test", "debug", DevNull())
	// Just verify no panic occurs
	WithTrace()
}

func TestGlobalLogger(t *testing.T) {
	assert.NotNil(t, Global)

	newLog := New("new-logger")
	SetGlobalLogger(newLog)
	assert.Equal(t, newLog, GetGlobalLogger())
}
