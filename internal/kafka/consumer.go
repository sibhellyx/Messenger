package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sibhellyx/Messenger/internal/config"
	"github.com/sibhellyx/Messenger/internal/models/entity"
)

type MessageProcessor interface {
	ProcessKafkaMessage(ctx context.Context, message entity.Message) error
}

type Consumer struct {
	reader    *kafka.Reader
	config    config.KafkaConfig
	processor MessageProcessor
	shutdown  chan struct{}
}

func NewConsumer(cfg config.KafkaConfig, processor MessageProcessor) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.TopicMessages,
		GroupID:        cfg.GroupID,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        time.Duration(cfg.MaxWaitTime) * time.Millisecond,
		CommitInterval: time.Second, // Автокоммит каждую секунду
		StartOffset:    kafka.FirstOffset,
	})

	consumer := &Consumer{
		reader:    reader,
		config:    cfg,
		processor: processor,
		shutdown:  make(chan struct{}),
	}

	slog.Info("Kafka consumer initialized",
		"brokers", cfg.Brokers,
		"topic", cfg.TopicMessages,
		"group_id", cfg.GroupID)

	return consumer
}

// Start
func (c *Consumer) Start(ctx context.Context) {
	slog.Info("Starting Kafka consumer")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Consumer stopped by context")
			c.reader.Close()
			return
		case <-c.shutdown:
			slog.Info("Consumer stopped by shutdown signal")
			c.reader.Close()
			return
		default:
			c.consumeMessage(ctx)
		}
	}
}

func (c *Consumer) consumeMessage(ctx context.Context) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		slog.Error("Failed to fetch message from Kafka", "error", err)
		time.Sleep(1 * time.Second)
		return
	}

	slog.Debug("Message received from Kafka",
		"topic", msg.Topic,
		"partition", msg.Partition,
		"offset", msg.Offset,
		"key", string(msg.Key),
		"size", len(msg.Value))

	var kafkaMessage entity.Message
	if err := json.Unmarshal(msg.Value, &kafkaMessage); err != nil {
		slog.Error("Failed to unmarshal Kafka message",
			"error", err,
			"raw_message", string(msg.Value))

		if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
			slog.Error("Failed to commit invalid message", "error", commitErr)
		}
		return
	}

	if err := c.processor.ProcessKafkaMessage(ctx, kafkaMessage); err != nil {
		slog.Error("Failed to process Kafka message",
			"message_id", kafkaMessage.ID,
			"chat_id", kafkaMessage.ChatID,
			"error", err)
		return
	}

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		slog.Error("Failed to commit message", "error", err)
		return
	}

	slog.Info("Message processed and committed successfully",
		"message_id", kafkaMessage.ID,
		"chat_id", kafkaMessage.ChatID,
		"user_id", kafkaMessage.UserID,
		"offset", msg.Offset)
}

// Stop consumer
func (c *Consumer) Stop() {
	close(c.shutdown)
}

// Close reader
func (c *Consumer) Close() error {
	return c.reader.Close()
}
