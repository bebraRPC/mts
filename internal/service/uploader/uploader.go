package uploader

import (
	"context"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/pkg/logger"
)

//структура принимает клиент бд, минио, контекст, logger и предоставляет 2 функции uploaadImage getimagebyID

// использовать в cmd/gate/main.go

//

type Upload struct {
	storageClient string
	minioClient   string
	ctx           context.Context
	logger        logger.Interface
}

func (u *Upload) UploadImage() error {
	// РЕАЛИЗОВАТь
	return nil
}

func (u *Upload) GetImageByID(id string) (domain.Image, error) {
	// РЕАЛИЗОВАТь
	return domain.Image{}, nil
}
