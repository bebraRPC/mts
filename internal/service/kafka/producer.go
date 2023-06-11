package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/menyasosali/mts/pkg/logger"
)

type ImageProducer struct {
	Topic    string
	Producer sarama.AsyncProducer
	Logger   logger.Interface
}

func NewImageProducer(brokers []string, topic string, logger logger.Interface) (*ImageProducer, error) {
	config := &sarama.Config{}
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	imageProducer := &ImageProducer{
		Topic:    topic,
		Producer: producer,
		Logger:   logger,
	}

	go imageProducer.handleSuccess()
	go imageProducer.handleErrors()

	return imageProducer, nil
}

func (p *ImageProducer) ProduceMessage(message []byte) {
	p.Producer.Input() <- &sarama.ProducerMessage{
		Topic: p.Topic,
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
