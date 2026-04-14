package sink

import (
	"context"
	"notification-service/internal/model"
)

type NotificationSink interface {
	Send(ctx context.Context, notification model.Notification) error
}
