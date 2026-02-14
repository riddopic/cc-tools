// Package notify provides desktop and audio notifications for hook events.
package notify

import (
	"fmt"
	"time"
)

// minutesPerHour is the number of minutes in one hour.
const minutesPerHour = 60

// QuietHours configuration for suppressing notifications.
type QuietHours struct {
	Enabled bool
	Start   string // "HH:MM" format.
	End     string // "HH:MM" format.
}

// IsActive returns true if the given time falls within quiet hours.
// Returns false if quiet hours are disabled.
func (qh QuietHours) IsActive(now time.Time) bool {
	if !qh.Enabled {
		return false
	}

	startH, startM, err := parseTime(qh.Start)
	if err != nil {
		return false
	}

	endH, endM, err := parseTime(qh.End)
	if err != nil {
		return false
	}

	nowMinutes := now.Hour()*minutesPerHour + now.Minute()
	startMinutes := startH*minutesPerHour + startM
	endMinutes := endH*minutesPerHour + endM

	if startMinutes <= endMinutes {
		// Same day range (e.g., 08:00 to 17:00).
		return nowMinutes >= startMinutes && nowMinutes < endMinutes
	}

	// Overnight range (e.g., 21:00 to 07:30).
	return nowMinutes >= startMinutes || nowMinutes < endMinutes
}

func parseTime(s string) (int, int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, 0, fmt.Errorf("parsing time %q: %w", s, err)
	}

	return h, m, nil
}
