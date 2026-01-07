package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	reader *kafka.Reader
	logger *zap.Logger
}

// ConsumerConfig holds configuration for Kafka consumer
type ConsumerConfig struct {
	Brokers     []string
	Topic       string
	GroupID     string
	Logger      *zap.Logger
	MinBytes    int // Minimum bytes to fetch
	MaxBytes    int // Maximum bytes to fetch
	MaxWait     time.Duration
	StartOffset int64 // kafka.FirstOffset or kafka.LastOffset
}

// MessageHandler processes a Kafka message
type MessageHandler func(ctx context.Context, message kafka.Message) error

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		Topic:       cfg.Topic,
		GroupID:     cfg.GroupID,
		MinBytes:    cfg.MinBytes,
		MaxBytes:    cfg.MaxBytes,
		MaxWait:     cfg.MaxWait,
		StartOffset: cfg.StartOffset,
		ReadBackoffMin: 100 * time.Millisecond,
		ReadBackoffMax: 1 * time.Second,
	})

	return &Consumer{
		reader: reader,
		logger: cfg.Logger,
	}
}

// Consume starts consuming messages and calls handler for each message
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	c.logger.Info("Starting consumer",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group-id", c.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer context cancelled, stopping")
			return ctx.Err()
		default:
			// Fetch message with timeout
			msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			message, err := c.reader.FetchMessage(msgCtx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded || err == context.Canceled {
					continue
				}
				c.logger.Error("Failed to fetch message", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}

			// Process message
			if err := c.processMessage(ctx, message, handler); err != nil {
				c.logger.Error("Failed to process message",
					zap.String("topic", message.Topic),
					zap.Int("partition", message.Partition),
					zap.Int64("offset", message.Offset),
					zap.Error(err),
				)
				// Don't commit on error - will retry
				continue
			}

			// Commit message after successful processing
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.logger.Error("Failed to commit message", zap.Error(err))
				// Continue processing - message will be reprocessed
			}
		}
	}
}

// processMessage processes a single message with retry logic
func (c *Consumer) processMessage(ctx context.Context, message kafka.Message, handler MessageHandler) error {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			c.logger.Warn("Retrying message processing",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
			)
			time.Sleep(backoff)
		}

		if err := handler(ctx, message); err == nil {
			return nil
		} else {
			lastErr = err
			c.logger.Warn("Message processing failed",
				zap.Int("attempt", attempt+1),
				zap.Error(err),
			)
		}
	}

	// After max retries, log and potentially send to DLQ
	c.logger.Error("Message processing failed after max retries",
		zap.String("topic", message.Topic),
		zap.Int("partition", message.Partition),
		zap.Int64("offset", message.Offset),
		zap.Error(lastErr),
	)

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// UnmarshalEvent unmarshals a Kafka message into an event
func UnmarshalEvent(message kafka.Message, event interface{}) error {
	if err := json.Unmarshal(message.Value, event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return nil
}

// GetEventType extracts event type from message headers
func GetEventType(message kafka.Message) string {
	for _, header := range message.Headers {
		if header.Key == "event-type" {
			return string(header.Value)
		}
	}
	return "unknown"
}

