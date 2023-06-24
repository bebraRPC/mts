package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/menyasosali/mts/config"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/pkg/logger"
)

type ImageProcessor interface {
	ProcessImage(ImgKafka) domain.ImgDescriptor
}

type ImageConsumer struct {
	Ctx       context.Context
	Logger    logger.Interface
	Processor ImageProcessor
	Consumer  sarama.ConsumerGroup
	Cfg       config.KafkaConfig
}

func NewImageConsumer(ctx context.Context, logger logger.Interface, processor ImageProcessor, cfg config.KafkaConfig,
) (*ImageConsumer, error) {
	config := sarama.NewConfig()

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	imageConsumer := &ImageConsumer{
		Ctx:       ctx,
		Logger:    logger,
		Processor: processor,
		Consumer:  consumer,
		Cfg:       cfg,
	}

	return imageConsumer, nil
}

type imageConsumerHandler struct {
	logger    logger.Interface
	processor ImageProcessor
}

func (c *ImageConsumer) Start() {
	go c.Consume()
}

func (c *ImageConsumer) Close() error {
	if c.Consumer != nil {
		err := c.Consumer.Close()
		if err != nil {
			return fmt.Errorf("failed to close Kafka consumer: %w", err)
		}
	}
	return nil
}

func (c *ImageConsumer) Consume() {
	consumerHandler := imageConsumerHandler{
		logger:    c.Logger,
		processor: c.Processor,
	}

	for c.Ctx.Err() != nil {
		err := c.Consumer.Consume(c.Ctx, []string{c.Cfg.Topic}, consumerHandler)
		if err != nil {
			c.Logger.Error(fmt.Sprintf("Failed to consume Kafka message: %v", err))
		}

	}
}
func (h imageConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h imageConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h imageConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		imgKafka, err := extractImageInfo(message.Value)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Failed to extract image info from Kafka message: %v", err))
			session.MarkMessage(message, "")
			continue
		}

		h.processor.ProcessImage(imgKafka)

		session.MarkMessage(message, "")
	}

	return nil
}

func extractImageInfo(messageValue []byte) (ImgKafka, error) {
	var data map[string]interface{}
	var imgKafka ImgKafka
	err := json.Unmarshal(messageValue, &data)
	if err != nil {
		return imgKafka, err
	}
	imgKafka.ID = data["id"].(string)
	imgKafka.Name = data["name"].(string)
	imgKafka.OriginalURL = data["originUrl"].(string)
	return imgKafka, nil
}
