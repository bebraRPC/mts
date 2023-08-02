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
	ProcessImage(context.Context, ImgKafka) domain.ImgDescriptor
}

type ImageConsumer struct {
	Logger    logger.Interface
	Processor ImageProcessor
	Consumer  sarama.Consumer
	Cfg       config.KafkaConfig
}

func NewImageConsumer(logger logger.Interface, processor ImageProcessor, cfg config.KafkaConfig,
) (*ImageConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	conn, err := sarama.NewConsumer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka conn: %w", err)
	}

	imageConsumer := &ImageConsumer{
		Logger:    logger,
		Processor: processor,
		Consumer:  conn,
		Cfg:       cfg,
	}

	return imageConsumer, nil
}

func (c *ImageConsumer) Start(ctx context.Context) {
	c.Logger.Info("Start - consumer.go - 51")
	go c.Consume(ctx)
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

func (c *ImageConsumer) Consume(ctx context.Context) {
	consumer, err := c.Consumer.ConsumePartition(c.Cfg.Topic, 0, sarama.OffsetOldest)
	if err != nil {
		c.Logger.Error(fmt.Sprintf("failed to consume partition: %w", err))
		return
	}

	c.Logger.Info("Consumer started")

	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				c.Logger.Error(fmt.Sprintf("Consumer error: %w", err))
				doneCh <- struct{}{}
			case msg := <-consumer.Messages():
				imgKafka, err := extractImageInfo(msg.Value, c.Logger)
				if err != nil {
					c.Logger.Error(fmt.Sprintf("Failed to extract image info from Kafka message: %v", err))
					continue
				}
				c.Processor.ProcessImage(ctx, imgKafka)

			}
		}
	}()

	<-doneCh
}

func extractImageInfo(messageValue []byte, logger logger.Interface) (ImgKafka, error) {
	var data map[string]string
	var imgKafka ImgKafka
	err := json.Unmarshal(messageValue, &data)

	if err != nil {
		return imgKafka, err
	}

	imgKafka.ID = data["imageID"]
	imgKafka.Name = data["name"]
	imgKafka.OriginalURL = data["originUrl"]
	return imgKafka, nil
}
