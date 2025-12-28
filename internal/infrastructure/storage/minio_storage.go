package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage implements StorageService using MinIO (S3-compatible)
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

// NewMinIOStorage creates a new MinIO storage service
func NewMinIOStorage(endpoint, accessKeyID, secretAccessKey, bucketName string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &MinIOStorage{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		useSSL:     useSSL,
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return storage, nil
}

// UploadFile uploads a file to MinIO
func (s *MinIOStorage) UploadFile(ctx context.Context, bucket string, key string, file io.Reader, contentType string) (string, error) {
	if bucket == "" {
		bucket = s.bucketName
	}

	// Upload the file
	_, err := s.client.PutObject(ctx, bucket, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the URL
	return s.GetFileURL(ctx, bucket, key)
}

// DownloadFile downloads a file from MinIO
func (s *MinIOStorage) DownloadFile(ctx context.Context, bucket string, key string) ([]byte, error) {
	if bucket == "" {
		bucket = s.bucketName
	}

	// Get object
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	// Read all data
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

// DeleteFile deletes a file from MinIO
func (s *MinIOStorage) DeleteFile(ctx context.Context, bucket string, key string) error {
	if bucket == "" {
		bucket = s.bucketName
	}

	err := s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileURL returns the URL to access a file
func (s *MinIOStorage) GetFileURL(ctx context.Context, bucket string, key string) (string, error) {
	if bucket == "" {
		bucket = s.bucketName
	}

	// Generate presigned URL (valid for 7 days)
	url, err := s.client.PresignedGetObject(ctx, bucket, key, 7*24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// FileExists checks if a file exists in MinIO
func (s *MinIOStorage) FileExists(ctx context.Context, bucket string, key string) (bool, error) {
	if bucket == "" {
		bucket = s.bucketName
	}

	_, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}
