package sink

import (
	"context"
	"github.com/lars-sto/notification-service/internal/model"
)

type NotificationSink interface {
	Send(ctx context.Context, notification model.Notification) error
}
