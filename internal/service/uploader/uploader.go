package uploader

import (
	"context"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/internal/service/db"
	"github.com/menyasosali/mts/internal/service/minio"
	"github.com/menyasosali/mts/pkg/logger"
)

//структура принимает клиент бд, минио, контекст, logger и предоставляет 2 функции uploaadImage getimagebyID

// использовать в cmd/gate/main.go

//

type Upload struct {
	Ctx           context.Context
	StorageClient *db.Storager
	MinioClient   *minio.ClientMinio
	Logger        logger.Interface
}

func NewUploader(ctx context.Context, storageClient *db.Storager, minioClient *minio.ClientMinio, logger logger.Interface) *Upload {
	return &Upload{
		Ctx:           ctx,
		StorageClient: storageClient,
		MinioClient:   minioClient,
		Logger:        logger,
	}
}

func (u *Upload) UploadImage() error {
	//fileURL, err := u.MinioClient.UploadFile()
	return nil
}

func (u *Upload) GetImageByID(id string) (domain.Image, error) {
	// РЕАЛИЗОВАТь
	return domain.Image{}, nil
}
