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
	config.Consumer.Return.Errors = true

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka group: %w", err)
	}

	imageConsumer := &ImageConsumer{
		Ctx:       ctx,
		Logger:    logger,
		Processor: processor,
		Consumer:  group,
		Cfg:       cfg,
	}

	return imageConsumer, nil
}

type imageConsumerHandler struct {
	logger    logger.Interface
	processor ImageProcessor
}

func (c *ImageConsumer) Start() {
	c.Logger.Info("Start - consumer.go - 51")
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
	c.Logger.Info("67 - consumer.go - Consume() - start")
	consumerHandler := imageConsumerHandler{
		logger:    c.Logger,
		processor: c.Processor,
	}

	c.Logger.Info("73 - consumer.go - Consume() - consumerHandler: ", consumerHandler)
	for c.Ctx.Done() == nil {
		err := c.Consumer.Consume(c.Ctx, []string{c.Cfg.Topic}, consumerHandler)
		c.Logger.Info("75 - consumer.go - Consume() - After c.Consume.Consume(): ", consumerHandler)
		if err != nil {
			c.Logger.Error(fmt.Sprintf("Failed to consume Kafka message: %v", err))
		}

	}
}
func (h imageConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h imageConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h imageConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	h.logger.Info(fmt.Sprintf("imgKafka - consumer.go - ConsumerClaim - start"))
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
