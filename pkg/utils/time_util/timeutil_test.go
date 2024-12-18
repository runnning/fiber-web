package time_util

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	t.Run("format date", func(t *testing.T) {
		date := time.Date(2024, 3, 14, 15, 0, 0, 0, time.UTC)
		result := FormatDate(date)
		expected := "2024-03-14"

		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestStartOfDay(t *testing.T) {
	t.Run("get start of day", func(t *testing.T) {
		input := time.Date(2024, 3, 14, 15, 30, 45, 0, time.UTC)
		result := StartOfDay(input)
		expected := time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)

		if !result.Equal(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestAge(t *testing.T) {
	t.Run("calculate age", func(t *testing.T) {
		birthDate := time.Now().AddDate(-30, 0, 0)
		age := Age(birthDate)

		if age != 30 {
			t.Errorf("Expected age 30, got %d", age)
		}
	})
}

func TestIsWeekend(t *testing.T) {
	t.Run("check weekend", func(t *testing.T) {
		saturday := time.Date(2024, 3, 16, 12, 0, 0, 0, time.UTC)
		sunday := time.Date(2024, 3, 17, 12, 0, 0, 0, time.UTC)
		monday := time.Date(2024, 3, 18, 12, 0, 0, 0, time.UTC)

		if !IsWeekend(saturday) {
			t.Error("Expected Saturday to be weekend")
		}
		if !IsWeekend(sunday) {
			t.Error("Expected Sunday to be weekend")
		}
		if IsWeekend(monday) {
			t.Error("Expected Monday not to be weekend")
		}
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "hours and minutes",
			duration: 90 * time.Minute,
			expected: "1小时30分钟",
		},
		{
			name:     "days and hours",
			duration: 25 * time.Hour,
			expected: "1天1小时",
		},
		{
			name:     "only minutes",
			duration: 45 * time.Minute,
			expected: "45分钟",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
