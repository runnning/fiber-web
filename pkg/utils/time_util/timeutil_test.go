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
			name:     "零时长",
			duration: 0,
			expected: "0分钟",
		},
		{
			name:     "负时长",
			duration: -90 * time.Minute,
			expected: "-1小时30分钟",
		},
		{
			name:     "秒级时长",
			duration: 45 * time.Second,
			expected: "45秒",
		},
		{
			name:     "分钟级时长",
			duration: 45 * time.Minute,
			expected: "45分钟",
		},
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
			name:     "天、小时和分钟",
			duration: 25*time.Hour + 30*time.Minute,
			expected: "1天1小时30分钟",
		},
		{
			name:     "整天",
			duration: 24 * time.Hour,
			expected: "1天",
		},
		{
			name:     "整小时",
			duration: 2 * time.Hour,
			expected: "2小时",
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

func TestStartOfQuarter(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "第一季度",
			input:    time.Date(2024, 2, 15, 12, 30, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "第二季度",
			input:    time.Date(2024, 5, 1, 12, 30, 0, 0, time.UTC),
			expected: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "第三季度",
			input:    time.Date(2024, 8, 31, 12, 30, 0, 0, time.UTC),
			expected: time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfQuarter(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("期望得到 %v，实际得到 %v", tt.expected, result)
			}
		})
	}
}

func TestEndOfQuarter(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "第一季度末",
			input:    time.Date(2024, 2, 15, 12, 30, 0, 0, time.UTC),
			expected: time.Date(2024, 3, 31, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "第二季度末",
			input:    time.Date(2024, 5, 1, 12, 30, 0, 0, time.UTC),
			expected: time.Date(2024, 6, 30, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfQuarter(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("期望得到 %v，实际得到 %v", tt.expected, result)
			}
		})
	}
}

func TestGetQuarter(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected int
	}{
		{"一月属于第一季度", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 1},
		{"四月属于第二季度", time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), 2},
		{"七月属于第三季度", time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC), 3},
		{"十二月属于第四季度", time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetQuarter(tt.input)
			if result != tt.expected {
				t.Errorf("期望得到 %v，实际得到 %v", tt.expected, result)
			}
		})
	}
}

func TestIsWorkday(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected bool
	}{
		{"周一是工作日", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{"周六不是工作日", time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), false},
		{"周日不是工作日", time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWorkday(tt.input)
			if result != tt.expected {
				t.Errorf("期望得到 %v，实际得到 %v", tt.expected, result)
			}
		})
	}
}

func TestWorkdaysBetween(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "一周内的工作日",
			start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // 周一
			end:      time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), // 周五
			expected: 5,
		},
		{
			name:     "跨周的工作日",
			start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // 周一
			end:      time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), // 下周一
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WorkdaysBetween(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("期望得到 %v，实际得到 %v", tt.expected, result)
			}
		})
	}
}

func TestToTimestamp(t *testing.T) {
	t.Run("转换为时间戳", func(t *testing.T) {
		input := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		result := ToTimestamp(input)
		expected := input.Unix()
		if result != expected {
			t.Errorf("期望得到 %v，实际得到 %v", expected, result)
		}
	})
}

func TestToMilliTimestamp(t *testing.T) {
	t.Run("转换为毫秒时间戳", func(t *testing.T) {
		input := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		result := ToMilliTimestamp(input)
		expected := input.UnixNano() / int64(time.Millisecond)
		if result != expected {
			t.Errorf("期望得到 %v，实际得到 %v", expected, result)
		}
	})
}

func TestIsBetween(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    time.Time
		expected bool
	}{
		{"等于开始时间", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{"在时间范围内", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), true},
		{"等于结束时间", time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), true},
		{"早于开始时间", time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), false},
		{"晚于结束时间", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBetween(tt.input, start, end)
			if result != tt.expected {
				t.Errorf("期望得到 %v，实际得到 %v", tt.expected, result)
			}
		})
	}
}

func TestNextWorkday(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "周五的下一个工作日是下周一",
			input:    time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "周一的下一个工作日是周二",
			input:    time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NextWorkday(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("期望得到 %v，实际得到 %v", tt.expected, result)
			}
		})
	}
}
