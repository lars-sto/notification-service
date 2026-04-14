package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lars-sto/notification-service/internal/httpapi"
	"github.com/lars-sto/notification-service/internal/service"
	"github.com/lars-sto/notification-service/internal/sink"
	"github.com/lars-sto/notification-service/internal/telemetry"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	metrics := telemetry.NewMetrics(registry)

	notificationSink := buildNotificationSink(logger)
	notificationService := service.NewNotificationService(notificationSink, metrics)
	handler := httpapi.NewHandler(notificationService)

	mux := http.NewServeMux()
	mux.HandleFunc("/notifications", handler.HandleNotifications)
	mux.HandleFunc("/healthz", handler.HandleHealthz)
	mux.HandleFunc("/readyz", handler.HandleReadyz)
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	var h http.Handler = mux
	h = httpapi.WithLogging(logger, h)
	h = httpapi.WithMetrics(metrics, h)
	h = httpapi.WithRecover(logger, h)
	h = httpapi.WithRequestID(h)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("starting server", "addr", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func buildNotificationSink(logger *slog.Logger) sink.NotificationSink {
	webhookURL := strings.TrimSpace(os.Getenv("TEAMS_WEBHOOK_URL"))
	if webhookURL == "" {
		logger.Info("using memory sink")
		return sink.NewMemorySink()
	}

	teamsSink, err := sink.NewTeamsSink(webhookURL)
	if err != nil {
		logger.Error("failed to initialize teams sink", "error", err)
		os.Exit(1)
	}

	logger.Info("using teams sink")
	return teamsSink
}
