package gate

import (
	"context"
	"fmt"
	"github.com/menyasosali/mts/config"
	"github.com/menyasosali/mts/internal/service/db"
	"github.com/menyasosali/mts/internal/service/filestorer"
	"github.com/menyasosali/mts/internal/service/kafka"
	"github.com/menyasosali/mts/internal/service/minio"
	"github.com/menyasosali/mts/internal/transport"
	"github.com/menyasosali/mts/pkg/httpserver"
	"github.com/menyasosali/mts/pkg/logger"
	"github.com/menyasosali/mts/pkg/postgres"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg *config.GateConfig) {
	ctx := context.Background()

	// Logger
	l := logger.NewLogger(cfg.Log.Level)

	// Postgres
	pg, err := postgres.New(cfg.Postgres.URL, postgres.MaxPoolSize(cfg.Postgres.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// MinIO
	minioConfig := config.MinioConfig{
		Endpoint:   cfg.Minio.Endpoint,
		AccessKey:  cfg.Minio.AccessKey,
		SecretKey:  cfg.Minio.SecretKey,
		BucketName: cfg.Minio.BucketName,
	}
	minioClient, err := minio.NewMinioClient(ctx, l, minioConfig)
	if err != nil {
		log.Fatal("Failed to create MinIO client:", err)
	}

	// Kafka Producer
	kafkaProducerConfig := config.KafkaConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	}
	kafkaProducer, err := kafka.NewImageProducer(ctx, l, kafkaProducerConfig)
	if err != nil {
		log.Fatal("Failed to create Kafka producer:", err)
	}
	defer kafkaProducer.Close()

	// Uploader
	fileStorer := filestorer.NewFileStorer(ctx, l, minioClient)

	// DB Store
	store := db.NewStore(ctx, l, pg)

	// Transport
	newTransport := transport.NewTransport(ctx, l, fileStorer, store, kafkaProducer)

	// HTTP Server
	httpServer := httpserver.NewServer(ctx, l, newTransport, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-stop:
		l.Info("worker - main.go - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("worker - main.go - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		log.Fatal("HTTP server shutdown error:", err)
	}
}
