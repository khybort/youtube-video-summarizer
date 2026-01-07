package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer handles Kafka message production
type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

// ProducerConfig holds configuration for Kafka producer
type ProducerConfig struct {
	Brokers []string
	Logger  *zap.Logger
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll, // Wait for all replicas
		Async:        false,             // Synchronous writes for reliability
		WriteTimeout: 10 * time.Second,
		BatchSize:    1, // Send immediately
		BatchTimeout: 10 * time.Millisecond,
		Compression:  kafka.Snappy, // Compress messages
	}

	return &Producer{
		writer: writer,
		logger: cfg.Logger,
	}
}

// Publish sends a message to a Kafka topic
func (p *Producer) Publish(ctx context.Context, topic string, key string, event interface{}) error {
	// Serialize event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(getEventType(event))},
			{Key: "content-type", Value: []byte("application/json")},
		},
	}

	// Write message
	if err := p.writer.WriteMessages(ctx, message); err != nil {
		p.logger.Error("Failed to publish message",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Info("Message published",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.String("event-type", getEventType(event)),
	)

	return nil
}

// PublishWithRetry publishes a message with retry logic
func (p *Producer) PublishWithRetry(ctx context.Context, topic string, key string, event interface{}, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := p.Publish(ctx, topic, key, event); err == nil {
			return nil
		} else {
			lastErr = err
			if i < maxRetries-1 {
				backoff := time.Duration(i+1) * time.Second
				p.logger.Warn("Retrying publish",
					zap.String("topic", topic),
					zap.Int("attempt", i+1),
					zap.Duration("backoff", backoff),
				)
				time.Sleep(backoff)
			}
		}
	}
	return fmt.Errorf("failed to publish after %d retries: %w", maxRetries, lastErr)
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// getEventType extracts event type from event struct
func getEventType(event interface{}) string {
	switch event.(type) {
	case *VideoCreatedEvent:
		return "video.created"
	case *TranscriptRequestedEvent:
		return "transcript.requested"
	case *EmbeddingRequestedEvent:
		return "embedding.requested"
	case *SimilarityRequestedEvent:
		return "similarity.requested"
	case *AnalysisCompletedEvent:
		return "analysis.completed"
	case *AnalysisFailedEvent:
		return "analysis.failed"
	default:
		return "unknown"
	}
}

