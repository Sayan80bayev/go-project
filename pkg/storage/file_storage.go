package storage

import (
	"mime/multipart"
)

type FileStorage interface {
	UploadFile(file multipart.File, header *multipart.FileHeader) (string, error)

	DeleteFileByURL(fileURL string) error
}
