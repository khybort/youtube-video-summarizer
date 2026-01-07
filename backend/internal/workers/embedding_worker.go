package workers

import (
	"context"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/internal/services/cost"
	"youtube-video-summarizer/backend/internal/services/embedding"
	"youtube-video-summarizer/backend/internal/services/provider"
	kafkapkg "youtube-video-summarizer/backend/pkg/kafka"
)

// EmbeddingWorker processes embedding generation requests from Kafka
type EmbeddingWorker struct {
	consumer         *kafkapkg.Consumer
	embeddingService *embedding.Service
	videoRepo        repository.VideoRepository
	videoEventService interface {
		PublishSimilarityRequested(ctx context.Context, videoID uuid.UUID, youtubeID, targetVideoID string, priority int) error
		PublishAnalysisCompleted(ctx context.Context, videoID uuid.UUID, youtubeID string, hasTranscript, hasSummary, hasEmbedding bool, duration int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	}
	logger *zap.Logger
}

// NewEmbeddingWorker creates a new EmbeddingWorker
func NewEmbeddingWorker(
	consumer *kafkapkg.Consumer,
	embeddingService *embedding.Service,
	videoRepo repository.VideoRepository,
	videoEventService interface {
		PublishSimilarityRequested(ctx context.Context, videoID uuid.UUID, youtubeID, targetVideoID string, priority int) error
		PublishAnalysisCompleted(ctx context.Context, videoID uuid.UUID, youtubeID string, hasTranscript, hasSummary, hasEmbedding bool, duration int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	},
	logger *zap.Logger,
) *EmbeddingWorker {
	return &EmbeddingWorker{
		consumer:          consumer,
		embeddingService:  embeddingService,
		videoRepo:         videoRepo,
		videoEventService: videoEventService,
		logger:            logger,
	}
}

// Start starts the embedding worker
func (w *EmbeddingWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, w.handleMessage)
}

// handleMessage processes an embedding request message
func (w *EmbeddingWorker) handleMessage(ctx context.Context, message kafkago.Message) error {
	// Parse event
	var event kafkapkg.EmbeddingRequestedEvent
	if err := kafkapkg.UnmarshalEvent(message, &event); err != nil {
		w.logger.Error("Failed to unmarshal embedding.requested event", zap.Error(err))
		return err
	}

	videoID, err := uuid.Parse(event.VideoID)
	if err != nil {
		w.logger.Error("Invalid video ID in event", zap.String("video_id", event.VideoID), zap.Error(err))
		return err
	}

	w.logger.Info("Processing embedding request",
		zap.String("video_id", event.VideoID),
		zap.String("youtube_id", event.YouTubeID),
		zap.Int("priority", event.Priority),
	)

	startTime := time.Now()

	// Get video from repository
	video, err := w.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		w.logger.Error("Failed to get video", zap.String("video_id", event.VideoID), zap.Error(err))
		_ = w.videoEventService.PublishAnalysisFailed(
			ctx,
			videoID,
			event.YouTubeID,
			"embedding",
			"video not found",
			false, // Not retryable
		)
		return err
	}

	// Generate embeddings
	_, err = w.embeddingService.GenerateVideoEmbeddings(ctx, video, event.TranscriptContent)
	if err != nil {
		w.logger.Error("Failed to generate embeddings",
			zap.String("video_id", event.VideoID),
			zap.Error(err),
		)
		
		// Publish failure event
		_ = w.videoEventService.PublishAnalysisFailed(
			ctx,
			videoID,
			event.YouTubeID,
			"embedding",
			err.Error(),
			true, // Retryable
		)
		return err
	}

	duration := int(time.Since(startTime).Seconds())
	w.logger.Info("Embeddings generated successfully",
		zap.String("video_id", event.VideoID),
		zap.Int("duration_seconds", duration),
	)

	// Publish similarity request event
	if err := w.videoEventService.PublishSimilarityRequested(
		ctx,
		videoID,
		event.YouTubeID,
		"", // Empty target = calculate with all videos
		event.Priority,
	); err != nil {
		w.logger.Error("Failed to publish similarity.requested event",
			zap.String("video_id", event.VideoID),
			zap.Error(err),
		)
	}

	// Publish analysis completed event
	_ = w.videoEventService.PublishAnalysisCompleted(
		ctx,
		videoID,
		event.YouTubeID,
		true,  // hasTranscript
		false, // hasSummary (can be generated separately)
		true,  // hasEmbedding
		duration,
	)

	return nil
}

// StartEmbeddingWorker starts an embedding worker with the given configuration
func StartEmbeddingWorker(
	ctx context.Context,
	brokers []string,
	groupID string,
	embeddingRepo repository.EmbeddingRepository,
	videoRepo repository.VideoRepository,
	providerFactory *provider.ProviderFactory,
	costService *cost.Service,
	videoEventService interface {
		PublishSimilarityRequested(ctx context.Context, videoID uuid.UUID, youtubeID, targetVideoID string, priority int) error
		PublishAnalysisCompleted(ctx context.Context, videoID uuid.UUID, youtubeID string, hasTranscript, hasSummary, hasEmbedding bool, duration int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	},
	logger *zap.Logger,
) error {
	// Create embedding service
	// Note: embedding service needs transcript service, but we don't have it here
	// For now, pass nil - the embedding service will handle it
	embeddingService := embedding.NewService(
		embeddingRepo,
		providerFactory,
		costService,
		nil, // transcriptService not available in worker
		logger,
	)

	// Create consumer
	consumer := kafkapkg.NewConsumer(kafkapkg.ConsumerConfig{
		Brokers:     brokers,
		Topic:       kafkapkg.TopicEmbeddingRequested,
		GroupID:     groupID,
		Logger:      logger,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: -1, // LastOffset
	})

	// Create and start worker
	worker := NewEmbeddingWorker(consumer, embeddingService, videoRepo, videoEventService, logger)
	return worker.Start(ctx)
}
