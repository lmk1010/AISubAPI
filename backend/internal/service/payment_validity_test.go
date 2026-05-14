package service

import "testing"

func TestPsComputeValidityDaysSupportsPluralUnits(t *testing.T) {
	tests := []struct {
		unit string
		days int
		want int
	}{
		{unit: "day", days: 3, want: 3},
		{unit: "days", days: 3, want: 3},
		{unit: "week", days: 2, want: 14},
		{unit: "weeks", days: 2, want: 14},
		{unit: "month", days: 1, want: 30},
		{unit: "months", days: 1, want: 30},
		{unit: "year", days: 1, want: 365},
		{unit: "years", days: 1, want: 365},
	}

	for _, tt := range tests {
		t.Run(tt.unit, func(t *testing.T) {
			if got := psComputeValidityDays(tt.days, tt.unit); got != tt.want {
				t.Fatalf("psComputeValidityDays(%d, %q) = %d, want %d", tt.days, tt.unit, got, tt.want)
			}
		})
	}
}
