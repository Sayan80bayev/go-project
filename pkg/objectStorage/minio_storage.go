package objectStorage

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"sync"

	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var logger = logging.GetLogger()

// prefix holds the base URL prefix extracted from config
var prefix string

// MinioConfig holds only MinIO-related settings
type MinioConfig struct {
	Host      string
	Port      string
	AccessKey string
	SecretKey string
	Bucket    string
}

// MinioStorage is a wrapper around MinIO client and config
type MinioStorage struct {
	client *minio.Client
	cfg    *MinioConfig
}

// Singleton instance + sync.Once
var (
	storageInstance *MinioStorage
	once            sync.Once
)

// InitPrefix initializes the prefix variable from MinioConfig
func InitPrefix(cfg *MinioConfig) {
	prefix = fmt.Sprintf("http://%s:%s/", cfg.Host, cfg.Port)
}

// newClient creates a new MinIO client (internal use only)
func newClient(cfg *MinioConfig) *minio.Client {
	endpoint := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		logger.Fatalf("Failed to connect to MinIO: %v", err) // stop app on failure
	}
	logger.Info("Successfully connected to MinIO")
	return minioClient
}

// NewMinioStorage creates a new MinioStorage instance
func NewMinioStorage(cfg *MinioConfig) *MinioStorage {
	InitPrefix(cfg)
	client := newClient(cfg)
	return &MinioStorage{
		client: client,
		cfg:    cfg,
	}
}

// GetMinioStorage returns a singleton instance of MinioStorage
func GetMinioStorage(cfg *MinioConfig) *MinioStorage {
	once.Do(func() {
		InitPrefix(cfg)
		client := newClient(cfg)
		storageInstance = &MinioStorage{
			client: client,
			cfg:    cfg,
		}
	})
	return storageInstance
}

// UploadFile uploads a file using MinioStorage
func (s *MinioStorage) UploadFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	if file == nil || header == nil {
		return "", errors.New("invalid file")
	}
	defer file.Close()

	objectName := header.Filename
	contentType := header.Header.Get("Content-Type")
	bucketName := s.cfg.Bucket

	_, err := s.client.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	fileURL := fmt.Sprintf("%s%s/%s", prefix, bucketName, objectName)
	return fileURL, nil
}

// DeleteFileByURL deletes a file by URL using MinioStorage
func (s *MinioStorage) DeleteFileByURL(fileURL string) error {
	if fileURL == "" {
		return errors.New("missing file_url parameter")
	}
	if !strings.HasPrefix(fileURL, prefix) {
		return errors.New("invalid file_url format")
	}

	path := strings.TrimPrefix(fileURL, prefix)
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return errors.New("invalid file_url format")
	}

	bucketName := parts[0]
	objectName := parts[1]

	err := s.client.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
