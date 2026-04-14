package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lars-sto/notification-service/internal/service"
	"github.com/lars-sto/notification-service/internal/sink"
)

func setupHandler() *Handler {
	memSink := sink.NewMemorySink()
	svc := service.NewNotificationService(memSink, nil)
	return NewHandler(svc)
}

func TestHandleNotifications_WarningForwarded(t *testing.T) {
	handler := setupHandler()

	body := `{
		"type": "Warning",
		"name": "Backup Failure",
		"description": "Database issue"
	}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestHandleNotifications_InfoNotForwarded(t *testing.T) {
	handler := setupHandler()

	body := `{
		"type": "Info",
		"name": "Status",
		"description": "All good"
	}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestHandleNotifications_InvalidJSON(t *testing.T) {
	handler := setupHandler()

	body := `invalid-json`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleNotifications_UnknownField(t *testing.T) {
	handler := setupHandler()

	body := `{
		"type": "Warning",
		"name": "Backup",
		"description": "fail",
		"extra": "not allowed"
	}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleNotifications_WrongMethod(t *testing.T) {
	handler := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleNotifications_EmptyBody(t *testing.T) {
	handler := setupHandler()

	req := httptest.NewRequest(http.MethodPost, "/notifications", nil)
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleNotifications_UnsupportedType(t *testing.T) {
	handler := setupHandler()

	body := `{
		"type": "Debug",
		"name": "Test",
		"description": "not supported"
	}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleNotifications_MultipleJSONObjects(t *testing.T) {
	handler := setupHandler()

	body := `{"type":"Warning","name":"A","description":"B"} {"type":"Warning"}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleNotifications_MissingField(t *testing.T) {
	handler := setupHandler()

	body := `{
		"type": "Warning",
		"name": "Test"
	}`

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.HandleNotifications(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleHealthz(t *testing.T) {
	handler := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.HandleHealthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleReadyz(t *testing.T) {
	handler := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.HandleReadyz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
