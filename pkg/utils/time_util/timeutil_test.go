package time_util

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	t.Run("格式化日期", func(t *testing.T) {
		date := time.Date(2024, 3, 14, 15, 0, 0, 0, time.UTC)
		result := FormatDate(date)
		expected := "2024-03-14"

		if result != expected {
			t.Errorf("期望得到 %v，实际得到 %v", expected, result)
		}
	})
}

func TestStartOfDay(t *testing.T) {
	t.Run("获取一天的开始时间", func(t *testing.T) {
		input := time.Date(2024, 3, 14, 15, 30, 45, 0, time.UTC)
		result := StartOfDay(input)
		expected := time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)

		if !result.Equal(expected) {
			t.Errorf("期望得到 %v，实际得到 %v", expected, result)
		}
	})
}

func TestAge(t *testing.T) {
	t.Run("计算年龄", func(t *testing.T) {
		birthDate := time.Now().AddDate(-30, 0, 0)
		age := Age(birthDate)

		if age != 30 {
			t.Errorf("期望年龄为 30，实际得到 %d", age)
		}
	})
}

func TestIsWeekend(t *testing.T) {
	t.Run("检查是否为周末", func(t *testing.T) {
		saturday := time.Date(2024, 3, 16, 12, 0, 0, 0, time.UTC)
		sunday := time.Date(2024, 3, 17, 12, 0, 0, 0, time.UTC)
		monday := time.Date(2024, 3, 18, 12, 0, 0, 0, time.UTC)

		if !IsWeekend(saturday) {
			t.Error("期望周六是周末")
		}
		if !IsWeekend(sunday) {
			t.Error("期望周日是周末")
		}
		if IsWeekend(monday) {
			t.Error("期望周一不是周末")
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
			name:     "小时和分钟",
			duration: 90 * time.Minute,
			expected: "1小时30分钟",
		},
		{
			name:     "天和小时",
			duration: 25 * time.Hour,
			expected: "1天1小时",
		},
		{
			name:     "仅分钟",
			duration: 45 * time.Minute,
			expected: "45分钟",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("期望得到 %v，实际得到 %v", tt.expected, result)
			}
		})
	}
}
