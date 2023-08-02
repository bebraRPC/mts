package transport

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/menyasosali/mts/internal/service/db"
	"github.com/menyasosali/mts/internal/service/kafka"
	"io"
	"net/http"

	"github.com/menyasosali/mts/internal/service/filestorer"
	"github.com/menyasosali/mts/pkg/logger"
)

//хендлеры для сервака

// service filestorer

// http получают картинку закидывают в upload, в бд, минио

type Transport struct {
	Logger     logger.Interface
	FileStorer filestorer.FileStorerInterface
	Store      db.StoreInterface
	Producer   *kafka.ImageProducer //Producer
}

type ImageResponse struct {
	ImageID     string `json:"imageID"`
	Name        string `json:"name"`
	OriginalURL string `json:"originalUrl"`
}

type ImageDescriptorResponse struct {
	ImageID     string `json:"imageID"`
	OriginalURL string `json:"originalUrl"`
	Img512      string `json:"img512"`
	Img256      string `json:"img256"`
	Img16       string `json:"img16"`
}

func NewTransport(logger logger.Interface, fileStorer filestorer.FileStorerInterface, store db.StoreInterface,
	producer *kafka.ImageProducer) *Transport {

	return &Transport{
		Logger:     logger,
		FileStorer: fileStorer,
		Store:      store,
		Producer:   producer,
	}
}

func (t *Transport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	router.Get("/images/upload", t.ShowUploadPageHandler)
	router.Post("/images/upload", t.UploadImageHandler)
	router.Get("/images/get/{id}", t.GetImageByIDHandler)

	router.ServeHTTP(w, r)
}

func (t *Transport) ShowUploadPageHandler(w http.ResponseWriter, r *http.Request) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Upload Image</title>
		</head>
		<body>
			<h1>Upload Image</h1>
			<form action="/images/upload" method="POST" enctype="multipart/form-data">
				<input type="file" name="image" accept="image/*">
				<input type="submit" value="Upload">
			</form>
		</body>
		</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (t *Transport) UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		t.Logger.Error("Failed to parse multipart form", err)
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}
	t.Logger.Info("92.. - producer.go - Parse - success")
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

	imgURL, err := t.FileStorer.UploadImage(r.Context(), imageBytes, filename)
	if err != nil {
		t.Logger.Error("Failed to upload image", err)
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	t.Logger.Info("117.. - producer.go - FileStorer Upload - success")

	imgID, err := t.Store.UploadImage(r.Context(), filename, imgURL)
	if err != nil {
		t.Logger.Error("Failed to save image to db", err)
		http.Error(w, "Failed to save image to db", http.StatusInternalServerError)
		return
	}

	response := ImageResponse{
		ImageID:     imgID,
		Name:        filename,
		OriginalURL: imgURL,
	}

	message, err := json.Marshal(response)
	if err != nil {
		t.Logger.Error("Failed to marshal response to JSON", err)
		http.Error(w, "Failed to marshal response to JSON", http.StatusInternalServerError)
		return
	}
	t.Logger.Info(fmt.Sprintf("138.. - producer.go - message: %s", message))

	// в producer кидаю response в topic
	err = t.Producer.ProduceMessage(r.Context(), message)
	if err != nil {
		t.Logger.Error("Failed to produce message to Kafka topic", err)
		http.Error(w, "Failed to produce message to Kafka topic", http.StatusInternalServerError)
		return
	}

	// в docker-compose при диплое образа kafka manager проверяем есть ли topic или при рестарте создаем topic из config

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (t *Transport) GetImageByIDHandler(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "id")
	if imageID == "" {
		http.Error(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	img, err := t.Store.GetImageByID(r.Context(), imageID)
	if err != nil {
		t.Logger.Error("Failed to get image from db", err)
		http.Error(w, "Failed to get image from db", http.StatusInternalServerError)
		return
	}

	//фикс
	response := ImageDescriptorResponse{
		ImageID:     imageID,
		OriginalURL: img.URL,
		Img512:      img.URL512,
		Img256:      img.URL256,
		Img16:       img.URL16,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
