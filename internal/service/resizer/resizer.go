package resizer

import (
	"bytes"
	"fmt"
	"github.com/menyasosali/mts/internal/domain"
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
	Consumer    *kafka.ImageConsumer
	ClientMinio *minio.ClientMinio
	Logger      logger.Interface
}

func NewResizer(consumer *kafka.ImageConsumer, client *minio.ClientMinio, logger logger.Interface) *Resizer {
	return &Resizer{
		Consumer:    consumer,
		ClientMinio: client,
		Logger:      logger,
	}
}

func (r *Resizer) Start() {
	go r.Consumer.Consume()
}

func (r *Resizer) ProcessImage(img domain.Image) {
	originalImageBytes, err := r.ClientMinio.DownloadFile(img.ID)
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to download original image: %v", err))
		return
	}

	originalImage, _, err := image.Decode(bytes.NewReader(originalImageBytes))
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to decode original image: %v", err))
		return
	}

	imageType := filepath.Ext(img.URL)

	resizedImage512, err := resizeTo(512, originalImage, imageType)
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to resize image: %v", err))
		return
	}

	resizedImage256, err := resizeTo(256, originalImage, imageType)
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to resize image: %v", err))
		return
	}

	resizedImage16, err := resizeTo(16, originalImage, imageType)
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to resize image: %v", err))
		return
	}

	resizedImage512URL, err := r.ClientMinio.UploadFile(resizedImage512, img.ID+"ID-512")
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to upload resized image to MinIO: %v", err))
		return
	}

	resizedImage256URL, err := r.ClientMinio.UploadFile(resizedImage256, img.ID+"ID-256")
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to upload resized image to MinIO: %v", err))
		return
	}

	resizedImage16URL, err := r.ClientMinio.UploadFile(resizedImage16, img.ID+"ID-16")
	if err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to upload resized image to MinIO: %v", err))
		return
	}

	img.URL512 = resizedImage512URL
	img.URL256 = resizedImage256URL
	img.URL16 = resizedImage16URL

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
