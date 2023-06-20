package main

import (
	"context"
	"fmt"
	"github.com/menyasosali/mts/cmd/gate/conf"
	"github.com/menyasosali/mts/internal/service/minio"
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

	pg, err := postgres.New(cfg.PG.URL)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	minioClient := minio.NewMinioClient(ctx, l)

}
