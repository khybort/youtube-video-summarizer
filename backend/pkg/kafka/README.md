# Kafka Integration

This package provides a comprehensive Kafka integration for the YouTube Video Summarizer project.

## Architecture

### Event-Driven Processing Pipeline

```
Video Created → video.created topic
    ↓
Transcript Worker → video.transcript.requested topic
    ↓
Embedding Worker → video.embedding.requested topic
    ↓
Similarity Worker → video.similarity.requested topic
    ↓
Analysis Completed → video.analysis.completed topic
```

## Components

### Producer (`producer.go`)
- **Purpose**: Publishes events to Kafka topics
- **Features**:
  - Synchronous writes for reliability
  - Automatic retry with exponential backoff
  - Message compression (Snappy)
  - Event type headers for routing
  - Required acks from all replicas

### Consumer (`consumer.go`)
- **Purpose**: Consumes messages from Kafka topics
- **Features**:
  - Consumer groups for load balancing
  - Automatic message retry (3 attempts)
  - Graceful error handling
  - Message commit after successful processing
  - Configurable batch sizes and timeouts

### Events (`events.go`)
- **Event Types**:
  - `VideoCreatedEvent`: Published when a new video is created
  - `TranscriptRequestedEvent`: Request transcript generation
  - `EmbeddingRequestedEvent`: Request embedding generation
  - `SimilarityRequestedEvent`: Request similarity calculation
  - `AnalysisCompletedEvent`: Published when analysis completes
  - `AnalysisFailedEvent`: Published when analysis fails

### Topics (`topics.go`)
- **Topic Names**:
  - `video.created`: New video events
  - `video.transcript.requested`: Transcript generation requests
  - `video.embedding.requested`: Embedding generation requests
  - `video.similarity.requested`: Similarity calculation requests
  - `video.analysis.completed`: Analysis completion events
  - `video.analysis.failed`: Analysis failure events
  - `video.dlq`: Dead letter queue for failed messages

## Configuration

### Environment Variables
```bash
KAFKA_BROKERS=localhost:9092,kafka2:9092
KAFKA_ENABLED=true
KAFKA_CONSUMER_GROUP=youtube-analyzer
```

### Consumer Groups
- `transcript-worker`: Processes transcript requests
- `embedding-worker`: Processes embedding requests
- `similarity-worker`: Processes similarity requests
- `dlq-processor`: Processes dead letter queue

## Usage

### Publishing Events

```go
producer := kafka.NewProducer(kafka.ProducerConfig{
    Brokers: []string{"localhost:9092"},
    Logger:  logger,
})

event := kafka.NewVideoCreatedEvent(...)
err := producer.Publish(ctx, kafka.TopicVideoCreated, videoID, event)
```

### Consuming Events

```go
consumer := kafka.NewConsumer(kafka.ConsumerConfig{
    Brokers:     []string{"localhost:9092"},
    Topic:       kafka.TopicTranscriptRequested,
    GroupID:     "transcript-worker",
    Logger:      logger,
    StartOffset: -1, // LastOffset
})

err := consumer.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
    // Process message
    return nil
})
```

## Error Handling

- **Retry Logic**: Automatic retry with exponential backoff (3 attempts)
- **Dead Letter Queue**: Failed messages after max retries are logged
- **Fallback**: If Kafka is unavailable, system falls back to direct processing

## Performance Considerations

- **Partitioning**: Topics are partitioned for parallel processing
- **Consumer Groups**: Multiple workers can process messages in parallel
- **Batch Processing**: Configurable batch sizes for throughput optimization
- **Compression**: Snappy compression reduces network overhead

## Monitoring

- All events are logged with structured logging (zap)
- Event types are included in message headers
- Processing duration is tracked in completion events
- Error events include stage and retryability information

