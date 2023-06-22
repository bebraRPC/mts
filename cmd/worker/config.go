package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Kafka kafkaConfig `yaml:"kafka"`
	Minio minioConfig `yaml:"minio"`
	Log   LogConfig   `yaml:"logger"`
}

type kafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group_id"`
}

type minioConfig struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	BucketName string `yaml:"bucket_name"`
}

type LogConfig struct {
	Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
}

func LoadConfig(env string) (*Config, error) {
	filename := fmt.Sprintf("env/%s/config.yaml", env)
	configPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for config file: %w", err)
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return config, nil
}
