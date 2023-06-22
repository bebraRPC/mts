package main

import (
	"context"
	"fmt"
	"github.com/menyasosali/mts/internal/service/db"
	"github.com/menyasosali/mts/internal/service/filestorer"
	"github.com/menyasosali/mts/internal/service/kafka"
	"github.com/menyasosali/mts/internal/service/kafka/cfg"
	"github.com/menyasosali/mts/internal/service/minio"
	"github.com/menyasosali/mts/internal/service/minio/cfg"
	"github.com/menyasosali/mts/internal/transport"
	"github.com/menyasosali/mts/pkg/httpserver"
	"github.com/menyasosali/mts/pkg/logger"
	"github.com/menyasosali/mts/pkg/postgres"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Minio    MinioConfig    `yaml:"minio"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Log      LogConfig      `yaml:"logger"`
	HTTP     HTTPConfig     `yaml:"http"`
}

type PostgresConfig struct {
	PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
	URL     string `env-required:"true"                 env:"PG_URL"`
}

type MinioConfig struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	BucketName string `yaml:"bucket_name"`
}

type LogConfig struct {
	Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

type HTTPConfig struct {
	Port string `yaml:"port"`
}

func LoadConfig(env string) (Config, error) {
	configFile := fmt.Sprintf("./cmd/gate/env/%s/config.yaml", env)
	file, err := os.Open(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	return config, nil
}

func main() {
	ctx := context.Background()

	// Load configuration
	env := "dev" // Set the desired environment
	cfg, err := LoadConfig(env)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Logger
	l := logger.NewLogger(cfg.Log.Level)

	pg, err := postgres.New(cfg.Postgres.URL, postgres.MaxPoolSize(cfg.Postgres.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// MinIO
	minioConfig := miniocfg.Config{
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
	kafkaProducerConfig := kafkacfg.ProducerConfig{
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
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		log.Fatal("HTTP server shutdown error:", err)
	}
}
