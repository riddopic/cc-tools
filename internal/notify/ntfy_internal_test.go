//go:build testmode

package notify

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNtfyNotifier_Timeout(t *testing.T) {
	t.Parallel()

	notifier := NewNtfyNotifier(NtfyConfig{
		Topic:  "test-topic",
		Server: "https://ntfy.sh",
	})

	assert.Greater(t, notifier.client.Timeout, time.Duration(0),
		"HTTP client must have a non-zero timeout")
}
