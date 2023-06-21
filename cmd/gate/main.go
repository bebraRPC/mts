package main

import (
	"context"
	"fmt"
	"github.com/menyasosali/mts/cmd/gate/conf"
	"github.com/menyasosali/mts/internal/service/db"
	"github.com/menyasosali/mts/internal/service/minio"
	miniocfg "github.com/menyasosali/mts/internal/service/minio/cfg"
	"github.com/menyasosali/mts/internal/service/uploader"
	"github.com/menyasosali/mts/internal/transport"
	"github.com/menyasosali/mts/pkg/logger"
	"github.com/menyasosali/mts/pkg/postgres"
	"time"
)

//иниц сервис

//тестовый upload файл для проверки работы сервисов

//listener, transport(ctx,router) route в transport/handlers.go

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := conf.Config{}

	l := logger.New(cfg.Log.Level)

	storageConn, err := postgres.New(cfg.PG.URL)
	if err != nil {
		l.Fatal(fmt.Errorf("gate - main.go - postgres.New: %w", err))
	}
	defer storageConn.Close()

	store := db.NewStore(ctx, l, storageConn)

	minioClient, err := minio.NewMinioClient(ctx, l, miniocfg.Config{})
	if err != nil {
		l.Fatal(fmt.Errorf("gate - main.go - minio.NewMinioClient: %w", err))
	}

	uploadService := uploader.NewUploader(ctx, l, store, minioClient)

	handlers := transport.NewTransport(ctx, l, uploadService)

}
