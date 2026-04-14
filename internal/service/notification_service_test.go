package service

import (
	"context"
	"errors"
	"testing"

	"notification-service/internal/model"
)

type stubSink struct {
	sent []model.Notification
	err  error
}

func (s *stubSink) Send(_ context.Context, notification model.Notification) error {
	if s.err != nil {
		return s.err
	}

	s.sent = append(s.sent, notification)
	return nil
}

func TestNotificationService_Process_WarningIsForwarded(t *testing.T) {
	sink := &stubSink{}
	service := NewNotificationService(sink, nil)

	notification := model.Notification{
		Type:        model.NotificationTypeWarning,
		Name:        "Backup Failure",
		Description: "The backup failed due to a database problem",
	}

	result, err := service.Process(context.Background(), notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Accepted {
		t.Fatalf("expected accepted to be true")
	}

	if !result.Forwarded {
		t.Fatalf("expected forwarded to be true")
	}

	if len(sink.sent) != 1 {
		t.Fatalf("expected 1 forwarded notification, got %d", len(sink.sent))
	}

	if sink.sent[0] != notification {
		t.Fatalf("expected forwarded notification to match input")
	}
}

func TestNotificationService_Process_InfoIsAcceptedButNotForwarded(t *testing.T) {
	sink := &stubSink{}
	service := NewNotificationService(sink, nil)

	notification := model.Notification{
		Type:        model.NotificationTypeInfo,
		Name:        "Daily Status",
		Description: "All systems are running normally",
	}

	result, err := service.Process(context.Background(), notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Accepted {
		t.Fatalf("expected accepted to be true")
	}

	if result.Forwarded {
		t.Fatalf("expected forwarded to be false")
	}

	if len(sink.sent) != 0 {
		t.Fatalf("expected 0 forwarded notifications, got %d", len(sink.sent))
	}
}

func TestNotificationService_Process_EmptyTypeReturnsInvalidNotification(t *testing.T) {
	sink := &stubSink{}
	service := NewNotificationService(sink, nil)

	notification := model.Notification{
		Name:        "Missing Type",
		Description: "This notification has no type",
	}

	_, err := service.Process(context.Background(), notification)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidNotification) {
		t.Fatalf("expected ErrInvalidNotification, got %v", err)
	}
}

func TestNotificationService_Process_UnsupportedTypeReturnsError(t *testing.T) {
	sink := &stubSink{}
	service := NewNotificationService(sink, nil)

	notification := model.Notification{
		Type:        model.NotificationType("Debug"),
		Name:        "Debug Event",
		Description: "This type is not supported",
	}

	_, err := service.Process(context.Background(), notification)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestNotificationService_Process_EmptyNameReturnsInvalidNotification(t *testing.T) {
	sink := &stubSink{}
	service := NewNotificationService(sink, nil)

	notification := model.Notification{
		Type:        model.NotificationTypeWarning,
		Name:        "   ",
		Description: "Missing real name",
	}

	_, err := service.Process(context.Background(), notification)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidNotification) {
		t.Fatalf("expected ErrInvalidNotification, got %v", err)
	}
}

func TestNotificationService_Process_EmptyDescriptionReturnsInvalidNotification(t *testing.T) {
	sink := &stubSink{}
	service := NewNotificationService(sink, nil)

	notification := model.Notification{
		Type:        model.NotificationTypeWarning,
		Name:        "Missing Description",
		Description: "   ",
	}

	_, err := service.Process(context.Background(), notification)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidNotification) {
		t.Fatalf("expected ErrInvalidNotification, got %v", err)
	}
}

func TestNotificationService_Process_SinkErrorIsReturned(t *testing.T) {
	expectedErr := errors.New("sink failed")

	sink := &stubSink{err: expectedErr}
	service := NewNotificationService(sink, nil)

	notification := model.Notification{
		Type:        model.NotificationTypeWarning,
		Name:        "Backup Failure",
		Description: "The backup failed due to a database problem",
	}

	_, err := service.Process(context.Background(), notification)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected sink error, got %v", err)
	}
}
