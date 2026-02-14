package notify_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/notify"
)

func TestNtfyNotifier_Send(t *testing.T) {
	t.Parallel()

	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(body, &received))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	notifier := notify.NewNtfyNotifier(notify.NtfyConfig{
		Topic:    "test-topic",
		Server:   srv.URL,
		Token:    "",
		Priority: 3,
	})

	err := notifier.Send(context.Background(), "Test Title", "Test message")
	require.NoError(t, err)

	assert.Equal(t, "test-topic", received["topic"])
	assert.Equal(t, "Test Title", received["title"])
	assert.Equal(t, "Test message", received["message"])
	assert.EqualValues(t, 3, received["priority"])
}

func TestNtfyNotifier_Send_WithToken(t *testing.T) {
	t.Parallel()

	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	notifier := notify.NewNtfyNotifier(notify.NtfyConfig{
		Topic:    "test-topic",
		Server:   srv.URL,
		Token:    "tk_mytoken",
		Priority: 0,
	})

	err := notifier.Send(context.Background(), "Title", "Body")
	require.NoError(t, err)
	assert.Equal(t, "Bearer tk_mytoken", authHeader)
}

func TestNtfyNotifier_Send_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	notifier := notify.NewNtfyNotifier(notify.NtfyConfig{
		Topic:    "test-topic",
		Server:   srv.URL,
		Token:    "",
		Priority: 0,
	})

	err := notifier.Send(context.Background(), "Title", "Body")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
