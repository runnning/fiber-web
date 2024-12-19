package str

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"
)

// RandomString 生成指定长度的随机字符串
func RandomString(length int) string {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// RandomBytes 生成指定长度的随机字节数组
func RandomBytes(length int) []byte {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

// CamelCase 将字符串转换为驼峰命名
// example: hello_world -> helloWorld
func CamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := false

	for i, v := range []rune(s) {
		if v >= 'A' && v <= 'Z' {
			v = unicode.ToLower(v)
		}
		if v == '_' || v == '-' || v == ' ' {
			capNext = true
		} else if capNext {
			if i > 0 {
				v = unicode.ToUpper(v)
			}
			capNext = false
			n.WriteRune(v)
		} else {
			n.WriteRune(v)
		}
	}
	return n.String()
}

// SnakeCase 将符串转换为蛇形命名
// example: helloWorld -> hello_world
func SnakeCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s) + 3)

	for i, v := range s {
		if i > 0 && unicode.IsUpper(v) {
			n.WriteRune('_')
		}
		n.WriteRune(unicode.ToLower(v))
	}
	return n.String()
}

// KebabCase 将字符串转换为kebab命名
// example: helloWorld -> hello-world
func KebabCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s) + 3)

	for i, v := range s {
		if i > 0 && unicode.IsUpper(v) {
			n.WriteRune('-')
		}
		n.WriteRune(unicode.ToLower(v))
	}
	return n.String()
}

// Truncate 截断字符串到指定长度,超出部分用...替换
func Truncate(s string, length int) string {
	if length <= 0 {
		return ""
	}

	if len(s) <= length {
		return s
	}

	return s[:length-3] + "..."
}

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsEmpty 判断字符串是否为空(包括空格)
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 判断字符串是否不为空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// DefaultIfEmpty 如果字符串为空则返回默认值
func DefaultIfEmpty(s string, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}
