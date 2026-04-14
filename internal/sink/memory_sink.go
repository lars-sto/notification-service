package sink

import (
	"context"
	"github.com/lars-sto/notification-service/internal/model"
	"sync"
)

type MemorySink struct {
	mu            sync.Mutex
	notifications []model.Notification
}

func NewMemorySink() *MemorySink {
	return &MemorySink{
		notifications: make([]model.Notification, 0),
	}
}

func (s *MemorySink) Send(_ context.Context, notification model.Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notifications = append(s.notifications, notification)
	return nil
}

// Notifications - helper to see what has been sent
func (s *MemorySink) Notifications() []model.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()

	copied := make([]model.Notification, len(s.notifications))
	copy(copied, s.notifications)
	return copied
}
