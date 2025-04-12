package str

import (
	"testing"
)

func TestRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
	}{
		{"长度为零", 0, 0},
		{"奇数长度", 7, 6}, // 向下取整
		{"偶数长度", 8, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandomString(tt.length)
			if len(got) != tt.want {
				t.Errorf("RandomString(%d) = %v, 期望长度 %v", tt.length, got, tt.want)
			}
		})
	}
}

func TestRandomBytes(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
	}{
		{"长度为零", 0, 0},
		{"正数长度", 16, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandomBytes(tt.length)
			if len(got) != tt.want {
				t.Errorf("RandomBytes(%d) = %v, 期望长度 %v", tt.length, got, tt.want)
			}
		})
	}
}

func TestCamelCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"空字符串", "", ""},
		{"单个单词", "hello", "hello"},
		{"下划线分隔", "hello_world", "helloWorld"},
		{"中划线分隔", "hello-world", "helloWorld"},
		{"空格分隔", "hello world", "helloWorld"},
		{"混合大小写", "Hello_World", "helloWorld"},
		{"多种分隔符", "hello_world-test space", "helloWorldTestSpace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CamelCase(tt.s); got != tt.want {
				t.Errorf("CamelCase(%q) = %v, 期望 %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"空字符串", "", ""},
		{"单个单词", "hello", "hello"},
		{"驼峰命名", "helloWorld", "hello_world"},
		{"多个单词", "helloWorldTest", "hello_world_test"},
		{"已是蛇形", "hello_world", "hello_world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnakeCase(tt.s); got != tt.want {
				t.Errorf("SnakeCase(%q) = %v, 期望 %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestKebabCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"空字符串", "", ""},
		{"单个单词", "hello", "hello"},
		{"驼峰命名", "helloWorld", "hello-world"},
		{"多个单词", "helloWorldTest", "hello-world-test"},
		{"已是短横线", "hello-world", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KebabCase(tt.s); got != tt.want {
				t.Errorf("KebabCase(%q) = %v, 期望 %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		length int
		want   string
	}{
		{"空字符串", "", 5, ""},
		{"负数长度", "hello", -1, ""},
		{"零长度", "hello", 0, ""},
		{"长度小于字符串", "hello", 3, "..."},
		{"长度等于字符串", "hello", 5, "hello"},
		{"长度大于字符串", "hello", 10, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.s, tt.length); got != tt.want {
				t.Errorf("Truncate(%q, %d) = %v, 期望 %v", tt.s, tt.length, got, tt.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"空字符串", "", ""},
		{"单个字符", "a", "a"},
		{"ASCII字符串", "hello", "olleh"},
		{"Unicode字符串", "你好世界", "界世好你"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.s); got != tt.want {
				t.Errorf("Reverse(%q) = %v, 期望 %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"空字符串", "", true},
		{"仅空格", " ", true},
		{"制表符和空格", "\t \n", true},
		{"非空字符串", "hello", false},
		{"带空格字符串", " hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.s); got != tt.want {
				t.Errorf("IsEmpty(%q) = %v, 期望 %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestDefaultIfEmpty(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		defaultValue string
		want         string
	}{
		{"空字符串", "", "默认值", "默认值"},
		{"仅空格", " ", "默认值", "默认值"},
		{"非空字符串", "hello", "默认值", "hello"},
		{"带空格字符串", " hello ", "默认值", " hello "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultIfEmpty(tt.s, tt.defaultValue); got != tt.want {
				t.Errorf("DefaultIfEmpty(%q, %q) = %v, 期望 %v", tt.s, tt.defaultValue, got, tt.want)
			}
		})
	}
}

// 基准测试
func BenchmarkCamelCase(b *testing.B) {
	s := "hello_world_test_string"
	for i := 0; i < b.N; i++ {
		CamelCase(s)
	}
}

func BenchmarkSnakeCase(b *testing.B) {
	s := "helloWorldTestString"
	for i := 0; i < b.N; i++ {
		SnakeCase(s)
	}
}

func BenchmarkKebabCase(b *testing.B) {
	s := "helloWorldTestString"
	for i := 0; i < b.N; i++ {
		KebabCase(s)
	}
}

func TestPascalCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"空字符串", "", ""},
		{"单个单词", "hello", "Hello"},
		{"下划线分隔", "hello_world", "HelloWorld"},
		{"中划线分隔", "hello-world", "HelloWorld"},
		{"空格分隔", "hello world", "HelloWorld"},
		{"混合大小写", "Hello_World", "HelloWorld"},
		{"多种分隔符", "hello_world-test space", "HelloWorldTestSpace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PascalCase(tt.s); got != tt.want {
				t.Errorf("CamelCase(%q) = %v, 期望 %v", tt.s, got, tt.want)
			}
		})
	}
}
