package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/menyasosali/mts/pkg/logger"
)

type ImageProcessor interface {
	ProcessImage(imageID int, originURL string)
}

type ImageConsumer struct {
	Ctx       context.Context
	Logger    logger.Interface
	Processor ImageProcessor
	Consumer  sarama.ConsumerGroup
	Topic     string
	Group     string
}

func NewImageConsumer(ctx context.Context, processor ImageProcessor, logger logger.Interface,
	brokers []string, topic string, group string, // ctx, logger ,imageporc ...
) (*ImageConsumer, error) { // в конф consumer cfg труктуру с

	config := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	imageConsumer := &ImageConsumer{
		Ctx:       ctx,
		Logger:    logger,
		Processor: processor,
		Consumer:  consumer,
		Topic:     topic,
		Group:     group,
	}

	return imageConsumer, nil
}

type imageConsumerHandler struct {
	processor ImageProcessor
	logger    logger.Interface
}

func (c *ImageConsumer) Consume() {
	consumerHandler := imageConsumerHandler{
		processor: c.Processor,
		logger:    c.Logger,
	}

	for c.Ctx.Err() != nil {
		err := c.Consumer.Consume(c.Ctx, []string{c.Topic}, consumerHandler)
		if err != nil {
			c.Logger.Error(fmt.Sprintf("Failed to consume Kafka message: %v", err))
		}
	}
}
func (h imageConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h imageConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h imageConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		imageID, originalURL, err := extractImageInfo(message.Value)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Failed to extract image info from Kafka message: %v", err))
			session.MarkMessage(message, "")
			continue
		}

		h.processor.ProcessImage(imageID, originalURL)

		session.MarkMessage(message, "")
	}

	return nil
}

func extractImageInfo(messageValue []byte) (int, string, error) {
	var data map[string]interface{}
	err := json.Unmarshal(messageValue, &data)
	if err != nil {
		return 0, "", err
	}
	imageID := data["imageID"].(int)
	originURL := data["originUrl"].(string)
	return imageID, originURL, nil
}
