package transport

import (
	"context"
	"encoding/json"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/internal/service/kafka"
	"io"
	"net/http"

	"github.com/menyasosali/mts/internal/service/uploader"
	"github.com/menyasosali/mts/pkg/logger"
)

//хендлеры для сервака

// service uploader

// http получают картинку закидывают в upload, в бд, минио

type Transport struct {
	Ctx      context.Context
	Logger   logger.Interface
	Uploader uploader.UploadInterface
}

func NewTransport(ctx context.Context, logger logger.Interface, uploader uploader.UploadInterface) *Transport {
	return &Transport{
		Ctx:      ctx,
		Logger:   logger,
		Uploader: uploader,
	}
}

func (t *Transport) UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		t.Logger.Error("Failed to read uploaded file", err)
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		t.Logger.Error("Failed to read image bytes", err)
		http.Error(w, "Failed to read image bytes", http.StatusInternalServerError)
		return
	}

	filename := header.Filename

	// Upload image
	imgURL, err := t.Uploader.UploadImage(imageBytes, filename)
	if err != nil {
		t.Logger.Error("Failed to upload image", err)
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	//фикс
	response := kafka.ImgKafka{
		OriginalURL: imgURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (t *Transport) GetImageByIDHandler(w http.ResponseWriter, r *http.Request) {
	imageID := r.URL.Query().Get("id")
	if imageID == "" {
		http.Error(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	img, err := t.Uploader.GetImageById(imageID)
	if err != nil {
		t.Logger.Error("Failed to get image by ID", err)
		http.Error(w, "Failed to get image by ID", http.StatusInternalServerError)
		return
	}
	if img == nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	//фикс
	response := domain.ImgDescriptor{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
