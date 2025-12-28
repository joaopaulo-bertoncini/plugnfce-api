package storage

import (
	"context"
	"io"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	// UploadFile uploads a file and returns the URL/path to access it
	UploadFile(ctx context.Context, bucket string, key string, file io.Reader, contentType string) (string, error)

	// DownloadFile downloads a file and returns its content
	DownloadFile(ctx context.Context, bucket string, key string) ([]byte, error)

	// DeleteFile deletes a file from storage
	DeleteFile(ctx context.Context, bucket string, key string) error

	// GetFileURL returns the URL to access a file
	GetFileURL(ctx context.Context, bucket string, key string) (string, error)

	// FileExists checks if a file exists in storage
	FileExists(ctx context.Context, bucket string, key string) (bool, error)
}

// UploadResult contains information about an uploaded file
type UploadResult struct {
	URL      string
	Key      string
	Bucket   string
	Size     int64
	MimeType string
}
