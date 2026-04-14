package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/lars-sto/notification-service/internal/model"
	"github.com/lars-sto/notification-service/internal/service"
)

type Handler struct {
	notificationService *service.NotificationService
}

func NewHandler(notificationService *service.NotificationService) *Handler {
	return &Handler{
		notificationService: notificationService,
	}
}

func (h *Handler) HandleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var notification model.Notification
	if err := decodeStrictJSON(w, r, &notification); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.notificationService.Process(r.Context(), notification)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidNotification):
			writeError(w, http.StatusBadRequest, "invalid notification")
		case errors.Is(err, service.ErrUnsupportedType):
			writeError(w, http.StatusBadRequest, "unsupported notification type")
		default:
			writeError(w, http.StatusInternalServerError, "failed to process notification")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) HandleReadyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body must not be empty")
		}
		return fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain exactly one JSON object")
	}

	return nil
}
