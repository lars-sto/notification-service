package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lars-sto/notification-service/internal/model"
)

var ErrTeamsWebhookFailed = errors.New("teams webhook request failed")

type TeamsSink struct {
	webhookURL string
	client     *http.Client
}

type teamsMessage struct {
	Text string `json:"text"`
}

func NewTeamsSink(webhookURL string) (*TeamsSink, error) {
	if strings.TrimSpace(webhookURL) == "" {
		return nil, errors.New("teams webhook url is required")
	}

	return &TeamsSink{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (s *TeamsSink) Send(ctx context.Context, notification model.Notification) error {
	payload := teamsMessage{
		Text: formatTeamsMessage(notification),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal teams payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create teams request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTeamsWebhookFailed, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: unexpected status code %d", ErrTeamsWebhookFailed, resp.StatusCode)
	}

	return nil
}

func formatTeamsMessage(notification model.Notification) string {
	return fmt.Sprintf(
		"[%s] %s\n%s",
		notification.Type,
		notification.Name,
		notification.Description,
	)
}
