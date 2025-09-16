package objectStorage

import (
	"context"
	"mime/multipart"
	"net/http"
)

type FileStorage interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
	DeleteFileByURL(ctx context.Context, fileURL string) error
	DownloadFile(w http.ResponseWriter, r *http.Request, fileURL string) error
}
