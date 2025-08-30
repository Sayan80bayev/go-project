package objectStorage

import (
    "context"
    "mime/multipart"
)

type FileStorage interface {
    UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
    DeleteFileByURL(ctx context.Context, fileURL string) error
}