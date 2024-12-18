package file_util

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetFileSize 获取文件大小
func GetFileSize(file *multipart.FileHeader) int64 {
	return file.Size
}

// GetFileExt 获取文件扩展名
func GetFileExt(fileName string) string {
	return strings.ToLower(filepath.Ext(fileName))
}

// CheckExist 检查文件是否存在
func CheckExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CheckPermission 检查文件权限
func CheckPermission(path string) bool {
	_, err := os.Stat(path)
	return os.IsPermission(err)
}

// CreateDirIfNotExist 如果目录不存在则创建
func CreateDirIfNotExist(path string) error {
	if !CheckExist(path) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

// SaveUploadedFile 保存上传的文件
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = CreateDirIfNotExist(filepath.Dir(dst)); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// DeleteFile 删除文件
func DeleteFile(path string) error {
	if CheckExist(path) {
		return os.Remove(path)
	}
	return nil
}

// ReadFile 读取文件内容
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile 写入文件内容
func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// GetFileInfo 获取文件信息
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// IsDir 检查是否是目录
func IsDir(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.IsDir()
	}
	return false
}

// ListFiles 列出目录下的所有文件
func ListFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err = CreateDirIfNotExist(filepath.Dir(dst)); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	if err := CopyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

// GetMimeType 获取文件MIME类型
func GetMimeType(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 读取文件前512字节用于判断文件类型
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// 使用http.DetectContentType检测MIME类型
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}
