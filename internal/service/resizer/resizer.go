package resizer

import (
	"bytes"
	"context"
	"fmt"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/internal/service/filestorer"
	"github.com/menyasosali/mts/internal/service/kafka"
	"github.com/menyasosali/mts/internal/service/minio"
	"github.com/menyasosali/mts/pkg/logger"
	"github.com/nfnt/resize"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"path/filepath"
)

// использует клиент кафки(consumer), minio

// использ в cmd/worker/main.go

type Resizer struct {
	Ctx         context.Context
	Logger      logger.Interface
	ClientMinio *minio.ClientMinio
	FileStorer  filestorer.FileStorerInterface
}

func NewResizer(ctx context.Context, logger logger.Interface,
	client *minio.ClientMinio, fileStorer filestorer.FileStorerInterface) *Resizer {

	return &Resizer{
		Ctx:         ctx,
		Logger:      logger,
		ClientMinio: client,
		FileStorer:  fileStorer,
	}
}

func (r *Resizer) ProcessImage(imgKafka kafka.ImgKafka) domain.ImgDescriptor {
	originalImageBytes, err := r.FileStorer.DownloadImage(imgKafka.ID)
	if err != nil {
		r.Logger.Error(err)
	}

	originalImage, _, err := image.Decode(bytes.NewReader(originalImageBytes))
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to decode original image: %v", err))
	}

	imageType := filepath.Ext(imgKafka.OriginalURL)

	resizedImage512, err := resizeTo(512, originalImage, imageType)
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to resize image: %v", err))
	}

	resizedImage256, err := resizeTo(256, originalImage, imageType)
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to resize image: %v", err))
	}

	resizedImage16, err := resizeTo(16, originalImage, imageType)
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to resize image: %v", err))
	}

	imgDescriptor := domain.ImgDescriptor{
		ID:   imgKafka.ID,
		Name: imgKafka.Name,
		URL:  imgKafka.OriginalURL,
	}

	imgDescriptor.URL512, err = r.FileStorer.UploadImage(resizedImage512, imgKafka.Name+"-512")
	if err != nil {
		r.Logger.Error(err)
	}

	imgDescriptor.URL256, err = r.ClientMinio.UploadFile(resizedImage256, imgKafka.Name+"-256")
	if err != nil {
		r.Logger.Error(err)
	}

	imgDescriptor.URL16, err = r.ClientMinio.UploadFile(resizedImage16, imgKafka.Name+"-16")
	if err != nil {
		r.Logger.Error(err)
	}

	return imgDescriptor
}

func resizeTo(width uint, originalImage image.Image, imageType string) ([]byte, error) {
	resizedImage := resize.Resize(width, 0, originalImage, resize.Lanczos3)
	switch imageType {
	case "jpg", "jpeg":
		var jpegBuffer bytes.Buffer
		err := jpeg.Encode(&jpegBuffer, resizedImage, nil)
		return jpegBuffer.Bytes(), err
	case "png":
		var pngBuffer bytes.Buffer
		err := png.Encode(&pngBuffer, resizedImage)
		return pngBuffer.Bytes(), err
	case "gif":
		var gifBuffer bytes.Buffer
		err := gif.Encode(&gifBuffer, resizedImage, nil)
		return gifBuffer.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}
