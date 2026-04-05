package service

import (
	"testing"
	"time"
)

func firstDay(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func endPtr(year int, month time.Month) *time.Time {
	t := firstDay(year, month)
	return &t
}

func TestMonthsOverlap(t *testing.T) {
	tests := []struct {
		name       string
		queryStart time.Time
		queryEnd   time.Time
		subStart   time.Time
		subEnd     *time.Time
		want       int
	}{
		{
			name:       "нет пересечения",
			queryStart: firstDay(2025, time.January),
			queryEnd:   firstDay(2025, time.March),
			subStart:   firstDay(2025, time.April),
			subEnd:     endPtr(2025, time.April),
			want:       0,
		},
		{
			name:       "один месяц внутри периода",
			queryStart: firstDay(2025, time.January),
			queryEnd:   firstDay(2025, time.March),
			subStart:   firstDay(2025, time.February),
			subEnd:     endPtr(2025, time.February),
			want:       1,
		},
		{
			name:       "подписка на весь период запроса",
			queryStart: firstDay(2025, time.January),
			queryEnd:   firstDay(2025, time.March),
			subStart:   firstDay(2025, time.January),
			subEnd:     endPtr(2025, time.March),
			want:       3,
		},
		{
			name:       "без даты окончания",
			queryStart: firstDay(2025, time.July),
			queryEnd:   firstDay(2025, time.December),
			subStart:   firstDay(2020, time.January),
			subEnd:     nil,
			want:       6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monthsOverlap(tt.queryStart, tt.queryEnd, tt.subStart, tt.subEnd)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
