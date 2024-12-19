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
		{"zero length", 0, 0},
		{"odd length", 7, 6}, // 向下取整
		{"even length", 8, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandomString(tt.length)
			if len(got) != tt.want {
				t.Errorf("RandomString(%d) = %v, want length %v", tt.length, got, tt.want)
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
		{"zero length", 0, 0},
		{"positive length", 16, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandomBytes(tt.length)
			if len(got) != tt.want {
				t.Errorf("RandomBytes(%d) = %v, want length %v", tt.length, got, tt.want)
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
		{"empty string", "", ""},
		{"single word", "hello", "hello"},
		{"snake case", "hello_world", "helloWorld"},
		{"kebab case", "hello-world", "helloWorld"},
		{"with spaces", "hello world", "helloWorld"},
		{"mixed case", "Hello_World", "helloWorld"},
		{"multiple delimiters", "hello_world-test space", "helloWorldTestSpace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CamelCase(tt.s); got != tt.want {
				t.Errorf("CamelCase(%q) = %v, want %v", tt.s, got, tt.want)
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
		{"empty string", "", ""},
		{"single word", "hello", "hello"},
		{"camel case", "helloWorld", "hello_world"},
		{"multiple words", "helloWorldTest", "hello_world_test"},
		{"already snake case", "hello_world", "hello_world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnakeCase(tt.s); got != tt.want {
				t.Errorf("SnakeCase(%q) = %v, want %v", tt.s, got, tt.want)
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
		{"empty string", "", ""},
		{"single word", "hello", "hello"},
		{"camel case", "helloWorld", "hello-world"},
		{"multiple words", "helloWorldTest", "hello-world-test"},
		{"already kebab case", "hello-world", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KebabCase(tt.s); got != tt.want {
				t.Errorf("KebabCase(%q) = %v, want %v", tt.s, got, tt.want)
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
		{"empty string", "", 5, ""},
		{"negative length", "hello", -1, ""},
		{"zero length", "hello", 0, ""},
		{"length less than string", "hello", 3, "..."},
		{"length equal to string", "hello", 5, "hello"},
		{"length greater than string", "hello", 10, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.s, tt.length); got != tt.want {
				t.Errorf("Truncate(%q, %d) = %v, want %v", tt.s, tt.length, got, tt.want)
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
		{"empty string", "", ""},
		{"single char", "a", "a"},
		{"ascii string", "hello", "olleh"},
		{"unicode string", "你好世界", "界世好你"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.s); got != tt.want {
				t.Errorf("Reverse(%q) = %v, want %v", tt.s, got, tt.want)
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
		{"empty string", "", true},
		{"space only", " ", true},
		{"tabs and spaces", "\t \n", true},
		{"non-empty string", "hello", false},
		{"string with spaces", " hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.s); got != tt.want {
				t.Errorf("IsEmpty(%q) = %v, want %v", tt.s, got, tt.want)
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
		{"empty string", "", "default", "default"},
		{"space only", " ", "default", "default"},
		{"non-empty string", "hello", "default", "hello"},
		{"string with spaces", " hello ", "default", " hello "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultIfEmpty(tt.s, tt.defaultValue); got != tt.want {
				t.Errorf("DefaultIfEmpty(%q, %q) = %v, want %v", tt.s, tt.defaultValue, got, tt.want)
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
