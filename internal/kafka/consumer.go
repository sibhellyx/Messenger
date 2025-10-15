package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
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
	dlqWriter *kafka.Writer
	config    config.KafkaConfig
	processor MessageProcessor
	shutdown  chan struct{}
	wg        sync.WaitGroup
	mu        sync.RWMutex
	isRunning bool
}

func NewConsumer(cfg config.KafkaConfig, processor MessageProcessor) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.TopicMessages,
		GroupID:        cfg.GroupID,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        time.Duration(cfg.MaxWaitTime) * time.Millisecond,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	dlqWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      cfg.Brokers,
		Topic:        cfg.TopicDLQ,
		Balancer:     &kafka.Hash{},
		MaxAttempts:  3,
		BatchTimeout: 50 * time.Millisecond,
	})

	consumer := &Consumer{
		reader:    reader,
		dlqWriter: dlqWriter,
		config:    cfg,
		processor: processor,
		shutdown:  make(chan struct{}),
		isRunning: false,
	}

	slog.Info("Kafka consumer initialized",
		"brokers", cfg.Brokers,
		"topic", cfg.TopicMessages,
		"group_id", cfg.GroupID)

	return consumer
}

// Start
func (c *Consumer) Start(ctx context.Context) {
	c.mu.Lock()
	if c.isRunning {
		c.mu.Unlock()
		slog.Warn("Consumer is already running")
		return
	}
	c.isRunning = true
	c.mu.Unlock()

	c.wg.Add(1)
	go c.run(ctx)

	slog.Info("Kafka consumer started")
}

func (c *Consumer) run(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Consumer stopped by context")
			return
		case <-c.shutdown:
			slog.Info("Consumer stopped by shutdown signal")
			return
		default:
			readCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)

			msg, err := c.reader.FetchMessage(readCtx)
			cancel()

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				if errors.Is(err, context.Canceled) {
					return
				}
				slog.Error("Failed to fetch message from Kafka", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			c.wg.Add(1)
			go c.processMessageAsync(ctx, msg)
		}
	}
}

func (c *Consumer) processMessageAsync(ctx context.Context, msg kafka.Message) {
	defer c.wg.Done()

	processingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := c.processMessageWithRetry(processingCtx, msg); err != nil {
		slog.Error("Failed to process message after retries",
			"offset", msg.Offset,
			"partition", msg.Partition,
			"error", err)
		c.handlePoisonPill(msg, err)
	}
}

func (c *Consumer) processMessageWithRetry(ctx context.Context, msg kafka.Message) error {
	var lastErr error

	for attempt := 1; attempt <= c.config.MaxRetry; attempt++ {
		if err := c.processSingleMessage(ctx, msg); err != nil {
			lastErr = err

			if attempt == c.config.MaxRetry {
				break
			}

			backoff := c.calculateBackoff(attempt)
			slog.Warn("Processing failed, retrying",
				"attempt", attempt,
				"backoff", backoff,
				"error", err)

			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	}

	return errors.New("max retry attempts exceeded: " + string(lastErr.Error()))
}

func (c *Consumer) processSingleMessage(ctx context.Context, message kafka.Message) error {
	var kafkaMessage entity.Message
	if err := json.Unmarshal(message.Value, &kafkaMessage); err != nil {
		if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
			slog.Error("Failed to commit invalid message", "error", commitErr)
		}
		return errors.New("invalid json message")
	}

	if err := c.processor.ProcessKafkaMessage(ctx, kafkaMessage); err != nil {
		slog.Error("Failed to process Kafka message",
			"message_id", kafkaMessage.ID,
			"chat_id", kafkaMessage.ChatID,
			"error", err)
		return errors.New("error process message")
	}

	if err := c.reader.CommitMessages(ctx, message); err != nil {
		slog.Error("commit error", "error", err)
		return errors.New("failed commit kafka messages")
	}

	slog.Debug("Message processed successfully",
		"message_id", kafkaMessage.ID,
		"offset", message.Offset)

	return nil

}

func (c *Consumer) handlePoisonPill(msg kafka.Message, processingErr error) {
	slog.Error("Handling poison pill message",
		"offset", msg.Offset,
		"partition", msg.Partition,
		"topic", msg.Topic,
		"error", processingErr,
		"raw_message", string(msg.Value))

	if c.config.TopicDLQ != "" {
		if err := c.sendToDLQ(msg, processingErr); err != nil {
			slog.Error("Failed to send message to DLQ", "error", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		slog.Error("Failed to commit poison pill message", "error", err)
	}

	slog.Warn("Poison pill message handled and committed",
		"offset", msg.Offset,
		"dlq_topic", c.config.TopicDLQ)
}

func (c *Consumer) sendToDLQ(msg kafka.Message, processingErr error) error {
	if c.config.TopicDLQ == "" {
		return errors.New("DLQ topic not configured")
	}

	dlqMessage := map[string]interface{}{
		"original_topic":     msg.Topic,
		"original_partition": msg.Partition,
		"original_offset":    msg.Offset,
		"original_key":       string(msg.Key),
		"original_value":     string(msg.Value),
		"error":              processingErr.Error(),
		"failed_at":          time.Now().UTC().Format(time.RFC3339),
		"headers":            msg.Headers,
	}

	dlqData, err := json.Marshal(dlqMessage)
	if err != nil {
		slog.Error("Failed to marshal DLQ message", "error", err)
		return errors.New("failed marshal dlq message")
	}

	dlqMsg := kafka.Message{
		Key:   []byte(fmt.Sprintf("dlq-%d-%d", msg.Partition, msg.Offset)),
		Value: dlqData,
		Headers: append(msg.Headers,
			kafka.Header{Key: "dlq-reason", Value: []byte(processingErr.Error())},
			kafka.Header{Key: "original-topic", Value: []byte(msg.Topic)},
			kafka.Header{Key: "original-offset", Value: []byte(fmt.Sprintf("%d", msg.Offset))},
		),
		Time: time.Now(),
	}

	if err := c.dlqWriter.WriteMessages(context.Background(), dlqMsg); err != nil {
		slog.Error("write to DLQ", "error", err)
		return errors.New("failed write to DLQ")
	}

	slog.Info("Message sent to DLQ",
		"dlq_topic", c.config.TopicDLQ,
		"original_offset", msg.Offset,
		"original_partition", msg.Partition)

	return nil
}

// Stop consumer
func (c *Consumer) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return
	}

	slog.Info("Stopping Kafka consumer...")
	close(c.shutdown)
	c.isRunning = false

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("Kafka consumer stopped gracefully")
	case <-time.After(30 * time.Second):
		slog.Warn("Kafka consumer stopped by timeout, some messages may not be processed")
	}
}

// Close reader and dlqWriter
func (c *Consumer) Close() error {
	c.Stop()

	var errs []error

	if err := c.reader.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close reader: %w", err))
	}

	if err := c.dlqWriter.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close DLQ writer: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing consumer: %v", errs)
	}

	slog.Info("Kafka consumer resources closed")
	return nil
}

// additional functions
func (c *Consumer) calculateBackoff(attempt int) time.Duration {
	if attempt <= 1 {
		initialBackoff := time.Duration(c.config.MaxWaitTime) * time.Millisecond
		if initialBackoff == 0 {
			initialBackoff = 100 * time.Millisecond // default value
		}
		return initialBackoff
	}

	// Expontianal : initial * 2^(attempt-1)
	initialBackoff := time.Duration(c.config.MaxWaitTime) * time.Millisecond
	if initialBackoff == 0 {
		initialBackoff = 100 * time.Millisecond
	}

	exponent := uint(attempt - 1)
	backoff := initialBackoff * (1 << exponent)

	// Max stopping
	maxBackoff := time.Duration(max(c.config.WriteTimeout, c.config.ReadTimeout)) * time.Second
	if maxBackoff == 0 {
		maxBackoff = 30 * time.Second
	}

	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return c.addJitter(backoff, 0.2)
}

func (c *Consumer) addJitter(duration time.Duration, jitterFactor float64) time.Duration {
	if jitterFactor <= 0 {
		return duration
	}

	jitterRange := float64(duration) * jitterFactor
	min := float64(duration) - jitterRange
	max := float64(duration) + jitterRange

	randomValue := min + rand.Float64()*(max-min)
	return time.Duration(randomValue)
}
