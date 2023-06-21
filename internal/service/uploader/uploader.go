package uploader

import (
	"context"
	"fmt"
	"github.com/menyasosali/mts/internal/service/db"
	"github.com/menyasosali/mts/internal/service/minio"
	"github.com/menyasosali/mts/pkg/logger"
)

//структура принимает клиент бд, минио, контекст, logger и предоставляет 2 функции uploaadImage getimagebyID

// использовать в cmd/gate/main.go

//

type UploadInterface interface {
	UploadImage([]byte, string) (string, error)
	GetImageByID(string) ([]byte, error)
}

type Uploader struct {
	Ctx           context.Context
	Logger        logger.Interface
	StorageClient *db.Store
	ClientMinio   *minio.ClientMinio
}

func NewUploader(ctx context.Context, logger logger.Interface, storageClient *db.Store, minioClient *minio.ClientMinio) *Uploader {
	return &Uploader{
		Ctx:           ctx,
		Logger:        logger,
		StorageClient: storageClient,
		ClientMinio:   minioClient,
	}
}

func (u *Uploader) UploadImage(imageBytes []byte, filename string) (string, error) {
	fileURL, err := u.ClientMinio.UploadFile(imageBytes, filename)

	if err != nil {
		u.Logger.Error(fmt.Sprintf("Failed to upload image to MinIO: %v", err))
		return "", fmt.Errorf("failed to upload image to MinIO: %w", err)
	}

	return fileURL, nil
}

func (u *Uploader) GetImageByID(imageID string) ([]byte, error) {
	originalImageBytes, err := u.ClientMinio.DownloadFile(imageID)
	if err != nil {
		errMsg := fmt.Errorf("failed to download original image: %v", err)
		u.Logger.Error(errMsg)
		return nil, errMsg
	}

	return originalImageBytes, nil
}
