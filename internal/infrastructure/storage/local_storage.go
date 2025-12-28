package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements StorageService using local filesystem
type LocalStorage struct {
	basePath   string
	publicURL  string
	bucketName string
}

// NewLocalStorage creates a new local filesystem storage service
func NewLocalStorage(basePath, publicURL, bucketName string) (*LocalStorage, error) {
	// Ensure base directory exists
	fullPath := filepath.Join(basePath, bucketName)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath:   basePath,
		publicURL:  strings.TrimSuffix(publicURL, "/"),
		bucketName: bucketName,
	}, nil
}

// UploadFile uploads a file to local filesystem
func (s *LocalStorage) UploadFile(ctx context.Context, bucket string, key string, file io.Reader, contentType string) (string, error) {
	if bucket == "" {
		bucket = s.bucketName
	}

	// Create directory structure if needed
	fullPath := filepath.Join(s.basePath, bucket, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return the URL
	return s.GetFileURL(ctx, bucket, key)
}

// DeleteFile deletes a file from local filesystem
func (s *LocalStorage) DeleteFile(ctx context.Context, bucket string, key string) error {
	if bucket == "" {
		bucket = s.bucketName
	}

	fullPath := filepath.Join(s.basePath, bucket, key)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileURL returns the URL to access a file
func (s *LocalStorage) GetFileURL(ctx context.Context, bucket string, key string) (string, error) {
	if bucket == "" {
		bucket = s.bucketName
	}

	// Return public URL
	url := fmt.Sprintf("%s/%s/%s", s.publicURL, bucket, key)
	return url, nil
}

// FileExists checks if a file exists in local filesystem
func (s *LocalStorage) FileExists(ctx context.Context, bucket string, key string) (bool, error) {
	if bucket == "" {
		bucket = s.bucketName
	}

	fullPath := filepath.Join(s.basePath, bucket, key)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// DownloadFile downloads a file from local filesystem
func (s *LocalStorage) DownloadFile(ctx context.Context, bucket string, key string) ([]byte, error) {
	if bucket == "" {
		bucket = s.bucketName
	}

	fullPath := filepath.Join(s.basePath, bucket, key)

	// Read file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", fullPath)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}
