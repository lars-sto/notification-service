package service

import (
	"context"
	"errors"
	"strings"

	"notification-service/internal/model"
	"notification-service/internal/sink"
	"notification-service/internal/telemetry"
)

var (
	ErrInvalidNotification = errors.New("invalid notification")
	ErrUnsupportedType     = errors.New("unsupported notification type")
	ErrSinkNotConfigured   = errors.New("notification sink is not configured")
)

type ProcessResult struct {
	Accepted  bool `json:"accepted"`
	Forwarded bool `json:"forwarded"`
}

type NotificationService struct {
	sink    sink.NotificationSink
	metrics *telemetry.Metrics
}

func NewNotificationService(notificationSink sink.NotificationSink, metrics *telemetry.Metrics) *NotificationService {
	return &NotificationService{
		sink:    notificationSink,
		metrics: metrics,
	}
}

func (s *NotificationService) Process(ctx context.Context, notification model.Notification) (ProcessResult, error) {
	if err := validateNotification(notification); err != nil {
		if s.metrics != nil {
			s.metrics.NotificationsRejected.Inc()
		}
		return ProcessResult{}, err
	}

	if s.metrics != nil {
		s.metrics.NotificationsReceived.WithLabelValues(string(notification.Type)).Inc()
	}

	switch notification.Type {
	case model.NotificationTypeWarning:
		if s.sink == nil {
			if s.metrics != nil {
				s.metrics.NotificationErrors.Inc()
			}
			return ProcessResult{}, ErrSinkNotConfigured
		}

		if err := s.sink.Send(ctx, notification); err != nil {
			if s.metrics != nil {
				s.metrics.NotificationErrors.Inc()
			}
			return ProcessResult{}, err
		}

		if s.metrics != nil {
			s.metrics.NotificationsForwarded.WithLabelValues(string(notification.Type)).Inc()
		}

		return ProcessResult{
			Accepted:  true,
			Forwarded: true,
		}, nil

	case model.NotificationTypeInfo:
		return ProcessResult{
			Accepted:  true,
			Forwarded: false,
		}, nil

	default:
		if s.metrics != nil {
			s.metrics.NotificationsRejected.Inc()
		}
		return ProcessResult{}, ErrUnsupportedType
	}
}

func validateNotification(notification model.Notification) error {
	if notification.Type == "" {
		return ErrInvalidNotification
	}
	if strings.TrimSpace(notification.Name) == "" {
		return ErrInvalidNotification
	}
	if strings.TrimSpace(notification.Description) == "" {
		return ErrInvalidNotification
	}
	return nil
}
