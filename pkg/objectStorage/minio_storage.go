package objectStorage

import (
	"context"
	"errors"
	"fmt"
	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"mime/multipart"
	"strings"
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

// InitPrefix initializes the prefix variable from MinioConfig
func InitPrefix(cfg *MinioConfig) {
	prefix = fmt.Sprintf("http://%s:%s/", cfg.Host, cfg.Port)
}

// Init creates and returns a MinIO client
func Init(cfg *MinioConfig) *minio.Client {
	endpoint := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		logger.Error(err)
	}
	logger.Info("Successfully connected to MinIO")
	return minioClient
}

// UploadFile uploads a file to MinIO and returns its URL
func UploadFile(file multipart.File, header *multipart.FileHeader, cfg *MinioConfig, minioClient *minio.Client) (string, error) {
	if file == nil || header == nil {
		return "", errors.New("invalid file")
	}
	defer file.Close()

	objectName := header.Filename
	contentType := header.Header.Get("Content-Type")
	bucketName := cfg.Bucket

	_, err := minioClient.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", errors.New("failed to upload file")
	}

	fileURL := fmt.Sprintf("%s%s/%s", prefix, bucketName, objectName)
	return fileURL, nil
}

// DeleteFileByURL deletes a file from MinIO using its URL
func DeleteFileByURL(fileURL string, minioClient *minio.Client) error {
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

	err := minioClient.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return errors.New("failed to delete file")
	}

	return nil
}

// MinioStorage is a wrapper around MinIO client and config
type MinioStorage struct {
	client *minio.Client
	cfg    *MinioConfig
}

// NewMinioStorage creates a new MinioStorage instance
func NewMinioStorage(client *minio.Client, cfg *MinioConfig) *MinioStorage {
	InitPrefix(cfg)
	return &MinioStorage{client: client, cfg: cfg}
}

// UploadFile uploads a file using MinioStorage
func (s *MinioStorage) UploadFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	return UploadFile(file, header, s.cfg, s.client)
}

// DeleteFileByURL deletes a file by URL using MinioStorage
func (s *MinioStorage) DeleteFileByURL(fileURL string) error {
	return DeleteFileByURL(fileURL, s.client)
}
