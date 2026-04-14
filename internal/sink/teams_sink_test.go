package sink

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"notification-service/internal/model"
)

func TestNewTeamsSink_EmptyURLReturnsError(t *testing.T) {
	_, err := NewTeamsSink("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTeamsSink_Send_SendsExpectedPayload(t *testing.T) {
	var (
		gotMethod      string
		gotContentType string
		gotBody        string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		gotBody = string(body)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sink, err := NewTeamsSink(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	notification := model.Notification{
		Type:        model.NotificationTypeWarning,
		Name:        "Backup Failure",
		Description: "Database issue",
	}

	err = sink.Send(context.Background(), notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}

	if gotContentType != "application/json" {
		t.Fatalf("expected application/json, got %s", gotContentType)
	}

	if !strings.Contains(gotBody, "[Warning] Backup Failure") {
		t.Fatalf("expected body to contain formatted title, got %s", gotBody)
	}

	if !strings.Contains(gotBody, "Database issue") {
		t.Fatalf("expected body to contain description, got %s", gotBody)
	}
}

func TestTeamsSink_Send_Non2xxReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	sink, err := NewTeamsSink(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	notification := model.Notification{
		Type:        model.NotificationTypeWarning,
		Name:        "Backup Failure",
		Description: "Database issue",
	}

	err = sink.Send(context.Background(), notification)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
