package logger

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestFieldConstructors(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected FieldType
	}{
		{"String", String("key", "value"), StringFieldType},
		{"Int", Int("key", 42), IntFieldType},
		{"Int64", Int64("key", 42), Int64FieldType},
		{"Float64", Float64("key", 3.14), Float64FieldType},
		{"Bool", Bool("key", true), BoolFieldType},
		{"ErrorField", ErrorField(errors.New("test error")), ErrorFieldType},
		{"Time", Time("key", time.Now()), TimeFieldType},
		{"Duration", Duration("key", time.Second), DurationFieldType},
		{"Binary", Binary("key", []byte("data")), BinaryFieldType},
		{"ByteString", ByteString("key", []byte("data")), ByteStringFieldType},
		{"Any", Any("key", "anything"), AnyFieldType},
		{"Skip", Skip(), SkipFieldType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Type() != tt.expected {
				t.Errorf("Expected field type %d, got %d", tt.expected, tt.field.Type())
			}
		})
	}
}

func TestFieldConversion(t *testing.T) {
	// 测试我们的 Field 转换为 zap.Field
	testCases := []struct {
		name     string
		field    Field
		expected zap.Field
	}{
		{
			"String field",
			String("test", "value"),
			zap.String("test", "value"),
		},
		{
			"Int field",
			Int("test", 42),
			zap.Int("test", 42),
		},
		{
			"Error field",
			ErrorField(errors.New("test error")),
			zap.Error(errors.New("test error")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			zapField := tc.field.toZapField()

			// 比较基本属性
			if zapField.Key != tc.expected.Key {
				t.Errorf("Expected key %s, got %s", tc.expected.Key, zapField.Key)
			}

			if zapField.Type != tc.expected.Type {
				t.Errorf("Expected type %d, got %d", tc.expected.Type, zapField.Type)
			}
		})
	}
}

func TestFromZapField(t *testing.T) {
	// 测试从 zap.Field 转换回我们的 Field
	zapFields := []zap.Field{
		zap.String("str", "value"),
		zap.Int("int", 42),
		zap.Bool("bool", true),
		zap.Error(errors.New("test error")),
	}

	for _, zapField := range zapFields {
		field := FromZapField(zapField)
		convertedBack := field.toZapField()

		// 验证往返转换保持一致性
		if convertedBack.Key != zapField.Key {
			t.Errorf("Round-trip conversion failed for key: expected %s, got %s",
				zapField.Key, convertedBack.Key)
		}

		if convertedBack.Type != zapField.Type {
			t.Errorf("Round-trip conversion failed for type: expected %d, got %d",
				zapField.Type, convertedBack.Type)
		}
	}
}

func TestConvertFields(t *testing.T) {
	fields := []Field{
		String("str", "value"),
		Int("int", 42),
		Bool("bool", true),
	}

	zapFields := convertFields(fields)

	if len(zapFields) != len(fields) {
		t.Errorf("Expected %d zap fields, got %d", len(fields), len(zapFields))
	}

	for i, field := range fields {
		if zapFields[i].Key != field.Key() {
			t.Errorf("Field %d: expected key %s, got %s",
				i, field.Key(), zapFields[i].Key)
		}
	}
}

func BenchmarkFieldConversion(b *testing.B) {
	field := String("benchmark", "test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = field.toZapField()
	}
}

func BenchmarkBatchConversion(b *testing.B) {
	fields := []Field{
		String("str", "value"),
		Int("int", 42),
		Bool("bool", true),
		ErrorField(errors.New("test")),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertFields(fields)
	}
}
