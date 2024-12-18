package time_util

import (
	"fmt"
	"time"
)

const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
	ShortDateTime  = "01-02 15:04"
)

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// FormatDateTime 格式化日期时间
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

// ParseDate 解析日期字符串
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse(DateFormat, dateStr)
}

// ParseDateTime 解析日期时间字符串
func ParseDateTime(dateTimeStr string) (time.Time, error) {
	return time.Parse(DateTimeFormat, dateTimeStr)
}

// StartOfDay 获取一天的开始时间
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取一天的结束时间
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek 获取一周的开始时间（周一）
func StartOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -int(weekday-1)))
}

// EndOfWeek 获取一周的结束时间（周日）
func EndOfWeek(t time.Time) time.Time {
	return EndOfDay(StartOfWeek(t).AddDate(0, 0, 6))
}

// StartOfMonth 获取月份的开始时间
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取月份的结束时间
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(StartOfMonth(t).AddDate(0, 1, -1))
}

// IsSameDay 判断两个时间是否是同一天
func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// IsWeekend 判断是否是周末
func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

// AddWorkDays 增加工作日
func AddWorkDays(t time.Time, days int) time.Time {
	for days > 0 {
		t = t.AddDate(0, 0, 1)
		if !IsWeekend(t) {
			days--
		}
	}
	return t
}

// DaysBetween 计算两个日期之间的天数
func DaysBetween(t1, t2 time.Time) int {
	duration := t2.Sub(t1)
	return int(duration.Hours() / 24)
}

// Age 计算年龄
func Age(birthDate time.Time) int {
	now := time.Now()
	years := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() ||
		(now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		years--
	}
	return years
}

// FormatDuration 格式化时间间隔
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%d天%d小时", days, hours)
		}
		return fmt.Sprintf("%d天", days)
	}
	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%d小时%d分钟", hours, minutes)
		}
		return fmt.Sprintf("%d小时", hours)
	}
	return fmt.Sprintf("%d分钟", minutes)
}

// RelativeTime 获取相对时间描述
func RelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	switch {
	case duration < time.Minute:
		return "刚刚"
	case duration < time.Hour:
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	case duration < 48*time.Hour:
		return "昨天"
	case duration < 72*time.Hour:
		return "前天"
	case duration < 7*24*time.Hour:
		return fmt.Sprintf("%d天前", int(duration.Hours()/24))
	default:
		return FormatDate(t)
	}
}
