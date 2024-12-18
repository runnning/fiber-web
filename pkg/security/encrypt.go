package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// MD5 计算字符串的 MD5 值
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256 计算字符串的 SHA256 值
func SHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// AESEncrypt AES 加密
func AESEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// AESDecrypt AES 解密
func AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// DESEncrypt DES 加密
func DESEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	plaintext = PKCS5Padding(plaintext, bs)
	if len(plaintext)%bs != 0 {
		return nil, errors.New("need a multiple of the blocksize")
	}

	ciphertext := make([]byte, len(plaintext))
	dst := ciphertext
	for len(plaintext) > 0 {
		block.Encrypt(dst, plaintext[:bs])
		plaintext = plaintext[bs:]
		dst = dst[bs:]
	}

	return ciphertext, nil
}

// DESDecrypt DES 解密
func DESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	if len(ciphertext)%bs != 0 {
		return nil, errors.New("input not full blocks")
	}

	plaintext := make([]byte, len(ciphertext))
	dst := plaintext
	for len(ciphertext) > 0 {
		block.Decrypt(dst, ciphertext[:bs])
		ciphertext = ciphertext[bs:]
		dst = dst[bs:]
	}

	return PKCS5UnPadding(plaintext), nil
}

// RC4Encrypt RC4 加密
func RC4Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))
	c.XORKeyStream(ciphertext, plaintext)
	return ciphertext, nil
}

// RC4Decrypt RC4 解密
func RC4Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	return RC4Encrypt(ciphertext, key) // RC4 加解密使用相同算法
}

// HMACSHA1 计算 HMAC-SHA1
func HMACSHA1(data []byte, key []byte) string {
	h := hmac.New(sha1.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// HMACSHA256 计算 HMAC-SHA256
func HMACSHA256(data []byte, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// HMACSHA512 计算 HMAC-SHA512
func HMACSHA512(data []byte, key []byte) string {
	h := hmac.New(sha512.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Base64Encode Base64 编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode Base64 解码
func Base64Decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

// Base64URLEncode Base64URL 编码
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode Base64URL 解码
func Base64URLDecode(str string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(str)
}

// PKCS5Padding PKCS5 填充
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS5UnPadding PKCS5 去除填充
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
