package model

type NotificationType string

const (
	NotificationTypeWarning NotificationType = "Warning"
	NotificationTypeInfo    NotificationType = "Info"
)

type Notification struct {
	Type        NotificationType `json:"type"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
}
