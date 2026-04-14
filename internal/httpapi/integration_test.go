package httpapi

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"notification-service/internal/service"
	"notification-service/internal/sink"
)

func TestIntegration_FullFlow_WarningForwarded(t *testing.T) {
	memSink := sink.NewMemorySink()
	svc := service.NewNotificationService(memSink, nil)
	handler := NewHandler(svc)

	logger := slog.Default()

	var h http.Handler = http.HandlerFunc(handler.HandleNotifications)
	h = WithLogging(logger, h)
	h = WithMetrics(nil, h)
	h = WithRecover(logger, h)
	h = WithRequestID(h)

	body := `{
		"type": "Warning",
		"name": "Backup Failure",
		"description": "Database issue"
	}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if len(memSink.Notifications()) != 1 {
		t.Fatalf("expected 1 forwarded notification, got %d", len(memSink.Notifications()))
	}
}

func TestIntegration_FullFlow_InfoNotForwarded(t *testing.T) {
	memSink := sink.NewMemorySink()
	svc := service.NewNotificationService(memSink, nil)
	handler := NewHandler(svc)

	logger := slog.Default()

	var h http.Handler = http.HandlerFunc(handler.HandleNotifications)
	h = WithLogging(logger, h)
	h = WithMetrics(nil, h)
	h = WithRecover(logger, h)
	h = WithRequestID(h)

	body := `{
		"type": "Info",
		"name": "Daily Status",
		"description": "All good"
	}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if len(memSink.Notifications()) != 0 {
		t.Fatalf("expected 0 forwarded notifications, got %d", len(memSink.Notifications()))
	}
}

func TestIntegration_RequestID_IsSet(t *testing.T) {
	memSink := sink.NewMemorySink()
	svc := service.NewNotificationService(memSink, nil)
	handler := NewHandler(svc)

	logger := slog.Default()

	var h http.Handler = http.HandlerFunc(handler.HandleNotifications)
	h = WithLogging(logger, h)
	h = WithMetrics(nil, h)
	h = WithRecover(logger, h)
	h = WithRequestID(h)

	body := `{
		"type": "Info",
		"name": "Status",
		"description": "ok"
	}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestIntegration_InvalidJSON_Returns400(t *testing.T) {
	memSink := sink.NewMemorySink()
	svc := service.NewNotificationService(memSink, nil)
	handler := NewHandler(svc)

	logger := slog.Default()

	var h http.Handler = http.HandlerFunc(handler.HandleNotifications)
	h = WithLogging(logger, h)
	h = WithMetrics(nil, h)
	h = WithRecover(logger, h)
	h = WithRequestID(h)

	body := `invalid-json`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestIntegration_PanicRecovered_Returns500(t *testing.T) {
	// Handler der absichtlich panict
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	logger := slog.Default()

	var h http.Handler = panicHandler
	h = WithLogging(logger, h)
	h = WithMetrics(nil, h)
	h = WithRecover(logger, h)
	h = WithRequestID(h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
