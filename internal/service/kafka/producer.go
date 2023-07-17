package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/menyasosali/mts/config"
	"github.com/menyasosali/mts/pkg/logger"
)

type ImageProducer struct {
	Ctx      context.Context
	Logger   logger.Interface
	Producer sarama.AsyncProducer
	Cfg      config.KafkaConfig
}

func NewImageProducer(ctx context.Context, logger logger.Interface, cfg config.KafkaConfig) (*ImageProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	logger.Info(fmt.Sprintf("Kafka producer SARAMA config - producer.go - 20: %s", config))

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	logger.Info(fmt.Sprintf("Kafka producer moi config - producer.go - 25: %s", cfg))

	imageProducer := &ImageProducer{
		Ctx:      ctx,
		Logger:   logger,
		Producer: producer,
		Cfg:      cfg,
	}

	return imageProducer, nil
}

func (p *ImageProducer) ProduceMessage(message []byte) error {
	messageInput := &sarama.ProducerMessage{
		Topic: p.Cfg.Topic,
		Value: sarama.ByteEncoder(message),
	}

	select {
	case p.Producer.Input() <- messageInput:
		p.Logger.Info(fmt.Sprintf("Successful send message from producer"))
		return nil
	case err := <-p.Producer.Errors():
		p.Logger.Error(fmt.Sprintf("Failed to produce Kafka message: %v", err))
		return err.Err
	case <-p.Ctx.Done():
		return p.Ctx.Err()
	}
}

func (p *ImageProducer) Close() error {
	return p.Producer.Close()
}
