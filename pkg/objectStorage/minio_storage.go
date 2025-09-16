package objectStorage

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioConfig holds only MinIO-related settings
type MinioConfig struct {
	Host       string
	Port       string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicHost string
}

// MinioStorage is a wrapper around MinIO client and config
type MinioStorage struct {
	client *minio.Client
	cfg    *MinioConfig
	prefix string
}

// Ensure MinioStorage implements FileStorage
var _ FileStorage = (*MinioStorage)(nil)

// NewMinioStorage creates a new MinioStorage instance
func NewMinioStorage(cfg *MinioConfig) (*MinioStorage, error) {
	if cfg == nil {
		return nil, errors.New("missing MinioConfig")
	}

	endpoint := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	logging.GetLogger().Info("✅ Successfully connected to MinIO")

	return &MinioStorage{
		client: client,
		cfg:    cfg,
		prefix: fmt.Sprintf("http://%s:%s/", cfg.PublicHost, cfg.Port),
	}, nil
}

// UploadFile uploads a file using MinioStorage
func (s *MinioStorage) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if file == nil || header == nil {
		return "", errors.New("invalid file")
	}
	defer file.Close()

	// Берём оригинальное имя файла
	originalName := header.Filename

	// Генерируем SHA1 хэш от имени
	h := sha1.New()
	h.Write([]byte(originalName))
	hashedName := hex.EncodeToString(h.Sum(nil))

	// Можно сохранить расширение файла, чтобы не терять формат
	// (например, .jpg, .png, .pdf и т.д.)
	ext := ""
	if idx := len(originalName) - 1 - len(header.Filename); idx >= 0 {
		// безопаснее использовать filepath.Ext
		ext = filepath.Ext(originalName)
	}
	objectName := hashedName + ext

	contentType := header.Header.Get("Content-Type")
	bucketName := s.cfg.Bucket

	_, err := s.client.PutObject(
		ctx,
		bucketName,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	fileURL := fmt.Sprintf("%s%s/%s", s.prefix, bucketName, objectName)
	return fileURL, nil
}

// DeleteFileByURL deletes a file by URL using MinioStorage
func (s *MinioStorage) DeleteFileByURL(ctx context.Context, fileURL string) error {
	if fileURL == "" {
		return errors.New("missing file_url parameter")
	}
	if !strings.HasPrefix(fileURL, s.prefix) {
		return errors.New("invalid file_url format")
	}

	path := strings.TrimPrefix(fileURL, s.prefix)
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return errors.New("invalid file_url format")
	}

	bucketName := parts[0]
	objectName := parts[1]

	err := s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DownloadFile for USER to download
func (s *MinioStorage) DownloadFile(w http.ResponseWriter, r *http.Request, fileURL string) error {
	if fileURL == "" {
		http.Error(w, "missing file_url parameter", http.StatusBadRequest)
		return errors.New("missing file_url parameter")
	}
	if !strings.HasPrefix(fileURL, s.prefix) {
		http.Error(w, "invalid file_url format", http.StatusBadRequest)
		return errors.New("invalid file_url format")
	}

	path := strings.TrimPrefix(fileURL, s.prefix)
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "invalid file_url format", http.StatusBadRequest)
		return errors.New("invalid file_url format")
	}

	bucketName := parts[0]
	objectName := parts[1]

	// Получаем объект из MinIO
	object, err := s.client.GetObject(r.Context(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get object: %v", err), http.StatusInternalServerError)
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Отправляем файл в ответ
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(objectName)))
	w.Header().Set("Content-Type", "application/octet-stream")

	if _, err := io.Copy(w, object); err != nil {
		http.Error(w, fmt.Sprintf("failed to stream file: %v", err), http.StatusInternalServerError)
		return fmt.Errorf("failed to stream file: %w", err)
	}

	return nil
}
