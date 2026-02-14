package notify_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/notify"
)

type mockSender struct {
	called bool
	err    error
}

func (m *mockSender) Send(_ context.Context, _, _ string) error {
	m.called = true
	return m.err
}

func TestMultiNotifier_Send_AllBackends(t *testing.T) {
	t.Parallel()

	s1 := &mockSender{called: false, err: nil}
	s2 := &mockSender{called: false, err: nil}

	multi := notify.NewMultiNotifier([]notify.Sender{s1, s2}, nil)
	err := multi.Send(context.Background(), "Title", "Body")
	require.NoError(t, err)

	assert.True(t, s1.called)
	assert.True(t, s2.called)
}

func TestMultiNotifier_Send_CollectsErrors(t *testing.T) {
	t.Parallel()

	s1 := &mockSender{called: false, err: errors.New("fail1")}
	s2 := &mockSender{called: false, err: errors.New("fail2")}

	multi := notify.NewMultiNotifier([]notify.Sender{s1, s2}, nil)
	err := multi.Send(context.Background(), "Title", "Body")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fail1")
	assert.Contains(t, err.Error(), "fail2")
}

func TestMultiNotifier_Send_QuietHours(t *testing.T) {
	t.Parallel()

	s1 := &mockSender{called: false, err: nil}
	qh := &notify.QuietHours{Enabled: true, Start: "00:00", End: "23:59"}

	multi := notify.NewMultiNotifier([]notify.Sender{s1}, qh)
	err := multi.Send(context.Background(), "Title", "Body")

	require.NoError(t, err)
	// During quiet hours, senders are NOT called.
	assert.False(t, s1.called)
}

func TestMultiNotifier_Send_NoSenders(t *testing.T) {
	t.Parallel()

	multi := notify.NewMultiNotifier(nil, nil)
	err := multi.Send(context.Background(), "Title", "Body")
	assert.NoError(t, err)
}
