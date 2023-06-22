package filestorer

import (
	"context"
	"fmt"
	"github.com/menyasosali/mts/internal/service/minio"
	"github.com/menyasosali/mts/pkg/logger"
)

//структура принимает клиент бд, минио, контекст, logger и предоставляет 2 функции uploaadImage getimagebyID

// использовать в cmd/gate/main.go

//

type FileStorerInterface interface {
	UploadImage([]byte, string) (string, error)
	DownloadImage(string) ([]byte, error)
}

type FileStorer struct {
	Ctx         context.Context
	Logger      logger.Interface
	ClientMinio *minio.ClientMinio
}

func NewFileStorer(ctx context.Context, logger logger.Interface, minioClient *minio.ClientMinio) *FileStorer {
	return &FileStorer{
		Ctx:         ctx,
		Logger:      logger,
		ClientMinio: minioClient,
	}
}

func (u *FileStorer) UploadImage(imageBytes []byte, filename string) (string, error) {
	fileURL, err := u.ClientMinio.UploadFile(imageBytes, filename)

	if err != nil {
		u.Logger.Error(fmt.Sprintf("Failed to upload image to MinIO: %v", err))
		return "", fmt.Errorf("failed to upload image to MinIO: %w", err)
	}

	return fileURL, nil
}

func (u *FileStorer) DownloadImage(imageID string) ([]byte, error) {
	originalImageBytes, err := u.ClientMinio.DownloadFile(imageID)
	if err != nil {
		errMsg := fmt.Errorf("failed to download original image: %v", err)
		u.Logger.Error(errMsg)
		return nil, errMsg
	}

	return originalImageBytes, nil
}
