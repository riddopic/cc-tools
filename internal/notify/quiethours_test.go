package notify_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/riddopic/cc-tools/internal/notify"
)

func TestQuietHours_IsActive(t *testing.T) {
	tests := []struct {
		name string
		qh   notify.QuietHours
		now  time.Time
		want bool
	}{
		{
			name: "10pm is quiet in overnight range 21:00-07:30",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "21:00",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 22, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "7am is quiet in overnight range 21:00-07:30",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "21:00",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 7, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "8am is not quiet in overnight range 21:00-07:30",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "21:00",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "3pm is not quiet in overnight range 21:00-07:30",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "21:00",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 15, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "exactly on start time is quiet",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "21:00",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 21, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "exactly on end time is NOT quiet",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "21:00",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 7, 30, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "disabled quiet hours always returns false",
			qh: notify.QuietHours{
				Enabled: false,
				Start:   "21:00",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 23, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "same-day range 08:00-17:00 at noon is quiet",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "08:00",
				End:     "17:00",
			},
			now:  time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "same-day range 08:00-17:00 at 7am is not quiet",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "08:00",
				End:     "17:00",
			},
			now:  time.Date(2025, 1, 15, 7, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "invalid start time format returns false",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "invalid",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 23, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "invalid end time format returns false",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "21:00",
				End:     "bad",
			},
			now:  time.Date(2025, 1, 15, 23, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "midnight 00:00 in overnight range is quiet",
			qh: notify.QuietHours{
				Enabled: true,
				Start:   "21:00",
				End:     "07:30",
			},
			now:  time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.qh.IsActive(tt.now)
			assert.Equal(t, tt.want, got)
		})
	}
}
