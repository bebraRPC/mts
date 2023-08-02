package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/menyasosali/mts/config"
	"github.com/menyasosali/mts/pkg/logger"
)

type ImageProducer struct {
	Logger   logger.Interface
	Producer sarama.SyncProducer
	Cfg      config.KafkaConfig
}

func NewImageProducer(logger logger.Interface, cfg config.KafkaConfig) (*ImageProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	logger.Info(fmt.Sprintf("Kafka producer SARAMA config - producer.go - 20: %s", config))

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	logger.Info(fmt.Sprintf("Kafka producer moi config - producer.go - 25: %s", cfg))

	imageProducer := &ImageProducer{
		Logger:   logger,
		Producer: producer,
		Cfg:      cfg,
	}

	return imageProducer, nil
}

func (p *ImageProducer) ProduceMessage(ctx context.Context, message []byte) error {
	messageInput := &sarama.ProducerMessage{
		Topic: p.Cfg.Topic,
		Value: sarama.ByteEncoder(message),
	}

	partition, offset, err := p.Producer.SendMessage(messageInput)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to send message to kafka: %s", err))
		return err
	}

	p.Logger.Info(fmt.Sprintf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", p.Cfg.Topic, partition, offset))
	return nil
}

func (p *ImageProducer) Close() error {
	return p.Producer.Close()
}
