package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	"github.com/sibhellyx/Messenger/internal/config"
)

type Producer struct {
	writer *kafka.Writer
	config config.KafkaConfig
}

type Message struct {
	Key     string
	Value   []byte
	Headers map[string]string
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.TopicMessages,
		Balancer:               &kafka.Hash{},
		MaxAttempts:            cfg.MaxRetry,
		BatchSize:              cfg.BatchSize,
		BatchTimeout:           time.Duration(cfg.LingerMS) * time.Millisecond,
		RequiredAcks:           kafka.RequiredAcks(cfg.RequiredAcks),
		Async:                  false,
		AllowAutoTopicCreation: true,
	}

	switch cfg.CompressionType {
	case "gzip":
		writer.Compression = compress.Gzip
	case "snappy":
		writer.Compression = compress.Snappy
	case "lz4":
		writer.Compression = compress.Lz4
	case "zstd":
		writer.Compression = compress.Zstd
	default:
		writer.Compression = compress.None
	}

	writer.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	writer.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second

	producer := &Producer{
		writer: writer,
		config: cfg,
	}

	slog.Info("Kafka producer initialized",
		"brokers", cfg.Brokers,
		"topic", cfg.TopicMessages,
		"max_retries", cfg.MaxRetry,
		"batch_size", cfg.BatchSize)

	return producer
}

// SendMessage send message Kafka
func (p *Producer) SendMessage(ctx context.Context, msg Message) error {
	kafkaMsg := kafka.Message{
		Key:   []byte(msg.Key),
		Value: msg.Value,
		Time:  time.Now(),
	}

	for key, value := range msg.Headers {
		kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	err := p.writer.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		slog.Error("Failed to send message to Kafka",
			"topic", p.config.TopicMessages,
			"key", msg.Key,
			"error", err)

		p.sendToDLQ(ctx, msg, err)
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	slog.Debug("Message sent to Kafka successfully",
		"topic", p.config.TopicMessages,
		"key", msg.Key,
		"message_size", len(msg.Value),
		"message", msg,
	)

	return nil
}

// SendJSON send JSON message
func (p *Producer) SendJSON(ctx context.Context, key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	msg := Message{
		Key:   key,
		Value: jsonData,
		Headers: map[string]string{
			"content-type": "application/json",
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		},
	}

	return p.SendMessage(ctx, msg)
}

// sendToDLQ отправляет сообщение в Dead Letter Queue
func (p *Producer) sendToDLQ(ctx context.Context, originalMsg Message, originalError error) {
	dlqMessage := map[string]interface{}{
		"original_topic": p.config.TopicMessages,
		"original_key":   originalMsg.Key,
		"original_value": string(originalMsg.Value),
		"error":          originalError.Error(),
		"failed_at":      time.Now().UTC().Format(time.RFC3339),
	}

	dlqBytes, err := json.Marshal(dlqMessage)
	if err != nil {
		slog.Error("Failed to marshal DLQ message", "error", err)
		return
	}

	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(p.config.Brokers...),
		Topic:    p.config.TopicDLQ,
		Balancer: &kafka.Hash{},
	}
	defer dlqWriter.Close()

	dlqMsg := kafka.Message{
		Key:   []byte(fmt.Sprintf("dlq_%s_%d", p.config.TopicMessages, time.Now().UnixNano())),
		Value: dlqBytes,
		Headers: []kafka.Header{
			{Key: "original_topic", Value: []byte(p.config.TopicMessages)},
			{Key: "error_type", Value: []byte(fmt.Sprintf("%T", originalError))},
		},
	}

	if err := dlqWriter.WriteMessages(ctx, dlqMsg); err != nil {
		slog.Error("Failed to send message to DLQ",
			"error", err,
			"original_error", originalError)
	}

}

func (p *Producer) SendJSONWithRetry(ctx context.Context, key string, value interface{}, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := p.SendJSON(ctx, key, value)
		if err == nil {
			return nil
		}

		lastErr = err
		slog.Warn("Failed to send message to Kafka, retrying...",
			"attempt", attempt,
			"max_retries", maxRetries,
			"key", key,
			"error", err)

		if attempt < maxRetries {
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				err := fmt.Errorf("context cancelled while retrying: %w", ctx.Err())
				slog.Warn("context cancelled", "err", err)
				return errors.New("failed send message")
			}
		}
	}
	err := fmt.Errorf("failed to send message after %d attempts: %w", maxRetries, lastErr)
	slog.Warn("Failed to send message to Kafka, retrying...",
		"key", key,
		"error", err)
	return errors.New("failed send message")
}

// close kafka writer
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// return stats of producer
func (p *Producer) GetStats() kafka.WriterStats {
	if p.writer != nil {
		return p.writer.Stats()
	}
	return kafka.WriterStats{}
}
