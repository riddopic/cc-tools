package notify

import (
	"context"
	"errors"
	"time"
)

// Sender sends a notification with a title and message body.
type Sender interface {
	Send(ctx context.Context, title, message string) error
}

// MultiNotifier composites multiple notification backends, respecting quiet hours.
type MultiNotifier struct {
	senders    []Sender
	quietHours *QuietHours
}

// NewMultiNotifier creates a notifier that fans out to all senders.
func NewMultiNotifier(senders []Sender, qh *QuietHours) *MultiNotifier {
	return &MultiNotifier{
		senders:    senders,
		quietHours: qh,
	}
}

// Send dispatches the notification to all backends. Errors are collected
// and returned as a joined error. Quiet hours suppress all notifications.
func (m *MultiNotifier) Send(ctx context.Context, title, message string) error {
	if m.quietHours != nil && m.quietHours.IsActive(time.Now()) {
		return nil
	}

	var errs []error

	for _, s := range m.senders {
		if err := s.Send(ctx, title, message); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
