package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/lars-sto/notification-service/internal/telemetry"

	"github.com/google/uuid"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type contextKey string

const requestIDKey contextKey = "request_id"
const requestIDHeader = "X-Request-ID"

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(requestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		w.Header().Set(requestIDHeader, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func WithLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := newResponseRecorder(w)

		next.ServeHTTP(recorder, r)

		logger.Info("http request completed",
			"request_id", GetRequestID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"status", recorder.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func WithMetrics(metrics *telemetry.Metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := newResponseRecorder(w)

		next.ServeHTTP(recorder, r)

		if metrics != nil {
			metrics.HTTPRequestTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(recorder.statusCode),
			).Inc()

			metrics.HTTPRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(time.Since(start).Seconds())
		}
	})
}

func WithRecover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered",
					"request_id", GetRequestID(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
					"panic", recovered,
					"stack_trace", string(debug.Stack()),
				)

				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
