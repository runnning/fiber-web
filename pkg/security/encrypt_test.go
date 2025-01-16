package security

import (
	"bytes"
	"strings"
	"testing"
)

func TestMD5(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "Hello World",
			input:    "Hello World",
			expected: "b10a8db164e0754105b7a99be72e3fe5",
		},
		{
			name:     "Complex string",
			input:    "Password123!@#",
			expected: "5dd51113856956e6d9cc84d7834600fc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MD5(tt.input); got != tt.expected {
				t.Errorf("MD5() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Hello World",
			input:    "Hello World",
			expected: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SHA256(tt.input); got != tt.expected {
				t.Errorf("SHA256() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAESEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       []byte
	}{
		{
			name:      "Empty string",
			plaintext: "",
			key:       []byte("0123456789abcdef"), // 16 bytes key
		},
		{
			name:      "Hello World",
			plaintext: "Hello World",
			key:       []byte("0123456789abcdef"),
		},
		{
			name:      "Long string",
			plaintext: "This is a long string that needs to be encrypted",
			key:       []byte("0123456789abcdef"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			encrypted, err := AESEncrypt([]byte(tt.plaintext), tt.key)
			if err != nil {
				t.Fatalf("AESEncrypt error: %v", err)
			}

			// 解密
			decrypted, err := AESDecrypt(encrypted, tt.key)
			if err != nil {
				t.Fatalf("AESDecrypt error: %v", err)
			}

			if string(decrypted) != tt.plaintext {
				t.Errorf("AES Encrypt/Decrypt failed: got %s, want %s", decrypted, tt.plaintext)
			}
		})
	}
}

func TestDESEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       []byte
	}{
		{
			name:      "Empty string",
			plaintext: "",
			key:       []byte("12345678"), // 8 bytes key
		},
		{
			name:      "Hello World",
			plaintext: "Hello World",
			key:       []byte("12345678"),
		},
		{
			name:      "Long string",
			plaintext: "This is a long string that needs to be encrypted",
			key:       []byte("12345678"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			encrypted, err := DESEncrypt([]byte(tt.plaintext), tt.key)
			if err != nil {
				t.Fatalf("DESEncrypt error: %v", err)
			}

			// 解密
			decrypted, err := DESDecrypt(encrypted, tt.key)
			if err != nil {
				t.Fatalf("DESDecrypt error: %v", err)
			}

			if string(decrypted) != tt.plaintext {
				t.Errorf("DES Encrypt/Decrypt failed: got %s, want %s", decrypted, tt.plaintext)
			}
		})
	}
}

func TestRC4EncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       []byte
	}{
		{
			name:      "Empty string",
			plaintext: "",
			key:       []byte("key"),
		},
		{
			name:      "Hello World",
			plaintext: "Hello World",
			key:       []byte("key"),
		},
		{
			name:      "Long string",
			plaintext: "This is a long string that needs to be encrypted",
			key:       []byte("key"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			encrypted, err := RC4Encrypt([]byte(tt.plaintext), tt.key)
			if err != nil {
				t.Fatalf("RC4Encrypt error: %v", err)
			}

			// 解密
			decrypted, err := RC4Decrypt(encrypted, tt.key)
			if err != nil {
				t.Fatalf("RC4Decrypt error: %v", err)
			}

			if string(decrypted) != tt.plaintext {
				t.Errorf("RC4 Encrypt/Decrypt failed: got %s, want %s", decrypted, tt.plaintext)
			}
		})
	}
}

func TestHMAC(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("Hello World")

	t.Run("HMAC-SHA1", func(t *testing.T) {
		result := HMACSHA1(data, key)
		if len(result) == 0 {
			t.Error("HMACSHA1 returned empty string")
		}
	})

	t.Run("HMAC-SHA256", func(t *testing.T) {
		result := HMACSHA256(data, key)
		if len(result) == 0 {
			t.Error("HMACSHA256 returned empty string")
		}
	})

	t.Run("HMAC-SHA512", func(t *testing.T) {
		result := HMACSHA512(data, key)
		if len(result) == 0 {
			t.Error("HMACSHA512 returned empty string")
		}
	})
}

func TestBase64(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Empty string",
			input: []byte(""),
		},
		{
			name:  "Hello World",
			input: []byte("Hello World"),
		},
		{
			name:  "Binary data",
			input: []byte{0, 1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Standard Base64
			encoded := Base64Encode(tt.input)
			decoded, err := Base64Decode(encoded)
			if err != nil {
				t.Fatalf("Base64Decode error: %v", err)
			}
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("Base64 Encode/Decode failed: got %v, want %v", decoded, tt.input)
			}

			// URL-safe Base64
			urlEncoded := Base64URLEncode(tt.input)
			urlDecoded, err := Base64URLDecode(urlEncoded)
			if err != nil {
				t.Fatalf("Base64URLDecode error: %v", err)
			}
			if !bytes.Equal(urlDecoded, tt.input) {
				t.Errorf("Base64URL Encode/Decode failed: got %v, want %v", urlDecoded, tt.input)
			}

			// 验证 URL 安全性
			if strings.Contains(urlEncoded, "+") || strings.Contains(urlEncoded, "/") {
				t.Error("Base64URL contains unsafe characters")
			}
		})
	}
}

func TestPKCS5Padding(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		blockSize int
	}{
		{
			name:      "Empty input",
			input:     []byte{},
			blockSize: 8,
		},
		{
			name:      "Full block",
			input:     []byte{1, 2, 3, 4, 5, 6, 7, 8},
			blockSize: 8,
		},
		{
			name:      "Partial block",
			input:     []byte{1, 2, 3},
			blockSize: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := PKCS5Padding(tt.input, tt.blockSize)
			if len(padded)%tt.blockSize != 0 {
				t.Errorf("PKCS5Padding result length %d is not multiple of block size %d", len(padded), tt.blockSize)
			}

			unpadded := PKCS5UnPadding(padded)
			if !bytes.Equal(unpadded, tt.input) {
				t.Errorf("PKCS5 Padding/Unpadding failed: got %v, want %v", unpadded, tt.input)
			}
		})
	}
}
