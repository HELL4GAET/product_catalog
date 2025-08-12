package file

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxFileSize  = 10 << 20
	AllowedTypes = "image/jpeg,image/png,application/pdf"
	KeyLength    = 16
)

type Storage interface {
	Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type FileService struct {
	sto Storage
}

func NewFileService(sto Storage) *FileService {
	return &FileService{sto: sto}
}

func (s *FileService) Upload(ctx context.Context, fh *multipart.FileHeader) (string, error) {
	if fh.Size > MaxFileSize {
		return "", errors.New("file size exceeds maximum allowed")
	}

	file, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	realType, err := detectContentType(file)
	if err != nil {
		return "", fmt.Errorf("detect content type: %w", err)
	}

	if !isAllowedType(realType) {
		return "", fmt.Errorf("file type not allowed: %s", realType)
	}

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("reset file pointer: %w", err)
	}

	key, err := generateSafeKey(filepath.Ext(fh.Filename))
	if err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}

	if err := s.sto.Upload(ctx, key, file, fh.Size, fh.Header.Get("Content-Type")); err != nil {
		return "", fmt.Errorf("upload file: %w", err)
	}

	url, err := s.sto.GetPresignedURL(ctx, key, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("get presigned URL: %w", err)
	}

	return url, nil
}

func generateSafeKey(ext string) (string, error) {
	buf := make([]byte, KeyLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	randomPart := hex.EncodeToString(buf)
	timestamp := time.Now().UnixNano()

	return fmt.Sprintf("%d_%s%s", timestamp, randomPart, sanitizeExt(ext)), nil
}

func sanitizeExt(ext string) string {
	ext = strings.ToLower(ext)
	if len(ext) > 10 {
		ext = ext[:10]
	}
	return ext
}

func detectContentType(file io.Reader) (string, error) {
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil && err != io.EOF {
		return "", err
	}
	return http.DetectContentType(buf), nil
}

func isAllowedType(contentType string) bool {
	allowed := strings.Split(AllowedTypes, ",")
	for _, t := range allowed {
		if strings.TrimSpace(t) == contentType {
			return true
		}
	}
	return false
}
