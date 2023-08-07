package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/menyasosali/mts/internal/service/db"
	"github.com/menyasosali/mts/internal/service/filestorer"
	"github.com/menyasosali/mts/internal/service/kafka"
	pb "github.com/menyasosali/mts/pkg/gen"
	"github.com/menyasosali/mts/pkg/logger"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net/http"
)

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

type Service struct {
	Logger     logger.Interface
	FileStorer filestorer.FileStorerInterface
	Store      db.StoreInterface
	Producer   *kafka.ImageProducer
	pb.UnimplementedGatewayServer
}

func NewService(log logger.Interface, fileStorer filestorer.FileStorerInterface, store db.StoreInterface,
	producer *kafka.ImageProducer) *Service {
	return &Service{
		Logger:     log,
		FileStorer: fileStorer,
		Store:      store,
		Producer:   producer,
	}
}

func (s *Service) GetUploadPage(context.Context, *emptypb.Empty) (*httpbody.HttpBody, error) {
	s.Logger.Info("GetUploadPage")
	return &httpbody.HttpBody{
		ContentType: "text/html",
		Data: []byte(`
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
	`),
	}, nil
}

func (s *Service) GetImageByID(ctx context.Context, req *pb.GetImageByIDRequest) (*pb.GetImageByIDResponse, error) {
	imageID := req.GetId()
	if imageID == "" {
		s.Logger.Error("Image ID is required")
		return nil, errors.New("image ID is required")
	}

	img, err := s.Store.GetImageByID(ctx, imageID)
	if err != nil {
		s.Logger.Error("Failed to get image from db", err)
		return nil, errors.New(fmt.Sprintf("Failed to get image from db: %v", err))
	}

	return &pb.GetImageByIDResponse{
		ImageID:     imageID,
		OriginalURL: img.URL,
		Img512:      img.URL512,
		Img256:      img.URL256,
		Img16:       img.URL16,
	}, nil
}

func (s *Service) UploadImageHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		s.Logger.Error("Failed to parse multipart form", err)
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}
	s.Logger.Info("92.. - producer.go - Parse - success")
	file, header, err := r.FormFile("image")
	if err != nil {
		s.Logger.Error("Failed to read uploaded file", err)
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		s.Logger.Error("Failed to read image bytes", err)
		http.Error(w, "Failed to read image bytes", http.StatusInternalServerError)
		return
	}

	filename := header.Filename

	imgURL, err := s.FileStorer.UploadImage(r.Context(), imageBytes, filename)
	if err != nil {
		s.Logger.Error("Failed to upload image", err)
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	s.Logger.Info("117.. - producer.go - FileStorer Upload - success")

	imgID, err := s.Store.UploadImage(r.Context(), filename, imgURL)
	if err != nil {
		s.Logger.Error("Failed to save image to db", err)
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
		s.Logger.Error("Failed to marshal response to JSON", err)
		http.Error(w, "Failed to marshal response to JSON", http.StatusInternalServerError)
		return
	}
	s.Logger.Info(fmt.Sprintf("138.. - producer.go - message: %s", message))

	// в producer кидаю response в topic
	err = s.Producer.ProduceMessage(r.Context(), message)
	if err != nil {
		s.Logger.Error("Failed to produce message to Kafka topic", err)
		http.Error(w, "Failed to produce message to Kafka topic", http.StatusInternalServerError)
		return
	}

	// в docker-compose при диплое образа kafka manager проверяем есть ли topic или при рестарте создаем topic из config

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
