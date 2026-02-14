package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// NtfyConfig configures the ntfy notification backend.
type NtfyConfig struct {
	Topic    string // required
	Server   string // default "https://ntfy.sh"
	Token    string // optional bearer token
	Priority int    // 1-5, default 3
}

// NtfyNotifier sends notifications via the ntfy.sh HTTP API.
type NtfyNotifier struct {
	config NtfyConfig
	client *http.Client
}

// NewNtfyNotifier creates a new ntfy notifier with the given configuration.
func NewNtfyNotifier(cfg NtfyConfig) *NtfyNotifier {
	if cfg.Server == "" {
		cfg.Server = "https://ntfy.sh"
	}

	if cfg.Priority == 0 {
		cfg.Priority = 3
	}

	return &NtfyNotifier{
		config: cfg,
		client: &http.Client{},
	}
}

// Send posts a notification to the configured ntfy topic.
func (n *NtfyNotifier) Send(ctx context.Context, title, message string) error {
	body := map[string]any{
		"topic":    n.config.Topic,
		"title":    title,
		"message":  message,
		"priority": n.config.Priority,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal ntfy payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.config.Server, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create ntfy request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if n.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.config.Token)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send ntfy notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}

	return nil
}
