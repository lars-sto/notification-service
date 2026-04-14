package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	NotificationsReceived  *prometheus.CounterVec
	NotificationsForwarded *prometheus.CounterVec
	NotificationsRejected  prometheus.Counter
	NotificationErrors     prometheus.Counter

	HTTPRequestTotal    *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
	m := &Metrics{
		NotificationsReceived: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_notifications_received_total",
				Help: "Total number of notifications received by type.",
			},
			[]string{"type"},
		),
		NotificationsForwarded: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_notifications_forwarded_total",
				Help: "Total number of notifications forwarded by type.",
			},
			[]string{"type"},
		),
		NotificationsRejected: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "notification_service_notifications_rejected_total",
				Help: "Total number of invalid or rejected notifications.",
			},
		),
		NotificationErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "notification_service_notification_errors_total",
				Help: "Total number of internal processing errors.",
			},
		),
		HTTPRequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "notification_service_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}

	registry.MustRegister(
		m.NotificationsReceived,
		m.NotificationsForwarded,
		m.NotificationsRejected,
		m.NotificationErrors,
		m.HTTPRequestTotal,
		m.HTTPRequestDuration,
	)

	return m
}
