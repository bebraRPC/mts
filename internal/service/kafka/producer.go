package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	producerCfg "github.com/menyasosali/mts/internal/service/kafka/cfg/producer"
	"github.com/menyasosali/mts/pkg/logger"
)

type ImageProducer struct {
	Ctx      context.Context
	Logger   logger.Interface
	Producer sarama.AsyncProducer
	Cfg      producerCfg.Config
}

func NewImageProducer(ctx context.Context, logger logger.Interface, cfg producerCfg.Config) (*ImageProducer, error) {
	config := &sarama.Config{}
	producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	imageProducer := &ImageProducer{
		Ctx:      ctx,
		Logger:   logger,
		Producer: producer,
		Cfg:      cfg,
	}

	go imageProducer.handleSuccess()
	go imageProducer.handleErrors()

	return imageProducer, nil
}

func (p *ImageProducer) ProduceMessage(message []byte) {
	p.Producer.Input() <- &sarama.ProducerMessage{
		Topic: p.Cfg.Topic,
		Value: sarama.ByteEncoder(message),
	}
}

func (p *ImageProducer) Close() error {
	return p.Producer.Close()
}

func (p *ImageProducer) handleSuccess() {
	for range p.Producer.Successes() {
		p.Logger.Info("Message sent successfully")
	}
}

func (p *ImageProducer) handleErrors() {
	for err := range p.Producer.Errors() {
		p.Logger.Error(fmt.Sprintf("Failed to produce Kafka message: %v", err))
	}
}
