package worker

import (
	"context"
	"fmt"
	"github.com/menyasosali/mts/config"
	"github.com/menyasosali/mts/internal/service/filestorer"
	"github.com/menyasosali/mts/internal/service/kafka"
	"github.com/menyasosali/mts/internal/service/minio"
	"github.com/menyasosali/mts/internal/service/resizer"
	"github.com/menyasosali/mts/pkg/logger"
	"log"
	"os"
	"os/signal"
	"time"
)

func Run(cfg *config.WorkerConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Logger
	l := logger.NewLogger(cfg.Log.Level)

	// Kafka Consumer
	kafkaConsumerConfig := config.KafkaConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
		GroupID: cfg.Kafka.GroupID,
	}

	// MinIO
	minioConfig := config.MinioConfig{
		Endpoint:   cfg.Minio.Endpoint,
		AccessKey:  cfg.Minio.AccessKey,
		SecretKey:  cfg.Minio.SecretKey,
		BucketName: cfg.Minio.BucketName,
	}
	minioClient, err := minio.NewMinioClient(ctx, l, minioConfig)
	if err != nil {
		l.Error(fmt.Errorf("failed to create MinIO client: %w", err))
	}

	// File Storer
	fileStorer := filestorer.NewFileStorer(ctx, l, minioClient)

	// Image Resizer
	processor := resizer.NewResizer(ctx, l, minioClient, fileStorer)

	// Kafka consumer
	kafkaConsumer, err := kafka.NewImageConsumer(ctx, l, processor, kafkaConsumerConfig)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}
	defer kafkaConsumer.Close()

	kafkaConsumer.Start()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case s := <-stop:
		l.Info("worker - main.go - signal: " + s.String())
	}

}
