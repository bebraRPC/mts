package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
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
