package workers

import (
	"context"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/internal/services/similarity"
	kafkapkg "youtube-video-summarizer/backend/pkg/kafka"
)

// SimilarityWorker processes similarity calculation requests from Kafka
type SimilarityWorker struct {
	consumer          *kafkapkg.Consumer
	similarityService *similarity.Service
	videoEventService interface {
		PublishAnalysisCompleted(ctx context.Context, videoID uuid.UUID, youtubeID string, hasTranscript, hasSummary, hasEmbedding bool, duration int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	}
	logger *zap.Logger
}

// NewSimilarityWorker creates a new SimilarityWorker
func NewSimilarityWorker(
	consumer *kafkapkg.Consumer,
	similarityService *similarity.Service,
	videoEventService interface {
		PublishAnalysisCompleted(ctx context.Context, videoID uuid.UUID, youtubeID string, hasTranscript, hasSummary, hasEmbedding bool, duration int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	},
	logger *zap.Logger,
) *SimilarityWorker {
	return &SimilarityWorker{
		consumer:          consumer,
		similarityService: similarityService,
		videoEventService: videoEventService,
		logger:            logger,
	}
}

// Start starts the similarity worker
func (w *SimilarityWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, w.handleMessage)
}

// handleMessage processes a similarity request message
func (w *SimilarityWorker) handleMessage(ctx context.Context, message kafkago.Message) error {
	// Parse event
	var event kafkapkg.SimilarityRequestedEvent
	if err := kafkapkg.UnmarshalEvent(message, &event); err != nil {
		w.logger.Error("Failed to unmarshal similarity.requested event", zap.Error(err))
		return err
	}

	videoID, err := uuid.Parse(event.VideoID)
	if err != nil {
		w.logger.Error("Invalid video ID in event", zap.String("video_id", event.VideoID), zap.Error(err))
		return err
	}

	w.logger.Info("Processing similarity request",
		zap.String("video_id", event.VideoID),
		zap.String("youtube_id", event.YouTubeID),
		zap.String("target_video_id", event.TargetVideoID),
		zap.Int("priority", event.Priority),
	)

	startTime := time.Now()

	// If target video ID is specified, calculate similarity with that specific video
	if event.TargetVideoID != "" {
		targetVideoID, err := uuid.Parse(event.TargetVideoID)
		if err != nil {
			w.logger.Error("Invalid target video ID in event", zap.String("target_video_id", event.TargetVideoID), zap.Error(err))
			_ = w.videoEventService.PublishAnalysisFailed(
				ctx,
				videoID,
				event.YouTubeID,
				"similarity",
				"invalid target video ID",
				false, // Not retryable
			)
			return err
		}

		// Calculate similarity with specific video
		_, err = w.similarityService.CalculateSimilarity(ctx, videoID, targetVideoID)
		if err != nil {
			w.logger.Error("Failed to calculate similarity with target video",
				zap.String("video_id", event.VideoID),
				zap.String("target_video_id", event.TargetVideoID),
				zap.Error(err),
			)
			_ = w.videoEventService.PublishAnalysisFailed(
				ctx,
				videoID,
				event.YouTubeID,
				"similarity",
				err.Error(),
				true, // Retryable
			)
			return err
		}
	} else {
		// Calculate similarity with all existing videos (find similar videos)
		_, err = w.similarityService.FindSimilarVideos(ctx, videoID, 10, 0.5)
		if err != nil {
			w.logger.Error("Failed to find similar videos",
				zap.String("video_id", event.VideoID),
				zap.Error(err),
			)
			_ = w.videoEventService.PublishAnalysisFailed(
				ctx,
				videoID,
				event.YouTubeID,
				"similarity",
				err.Error(),
				true, // Retryable
			)
			return err
		}
	}

	duration := int(time.Since(startTime).Seconds())
	w.logger.Info("Similarity calculation completed successfully",
		zap.String("video_id", event.VideoID),
		zap.Int("duration_seconds", duration),
	)

	// Publish analysis completed event (similarity is part of the analysis)
	_ = w.videoEventService.PublishAnalysisCompleted(
		ctx,
		videoID,
		event.YouTubeID,
		true, // hasTranscript (assumed to be true at this stage)
		false, // hasSummary (can be generated separately)
		true,  // hasEmbedding (assumed to be true at this stage)
		duration,
	)

	return nil
}

// StartSimilarityWorker starts a similarity worker with the given configuration
func StartSimilarityWorker(
	ctx context.Context,
	brokers []string,
	groupID string,
	similarityRepo repository.SimilarityRepository,
	embeddingRepo repository.EmbeddingRepository,
	videoRepo repository.VideoRepository,
	videoEventService interface {
		PublishAnalysisCompleted(ctx context.Context, videoID uuid.UUID, youtubeID string, hasTranscript, hasSummary, hasEmbedding bool, duration int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	},
	logger *zap.Logger,
) error {
	// Create similarity service (YouTube client not needed in worker, only for handler)
	similarityService := similarity.NewService(similarityRepo, embeddingRepo, videoRepo, nil, logger)

	// Create consumer
	consumer := kafkapkg.NewConsumer(kafkapkg.ConsumerConfig{
		Brokers:     brokers,
		Topic:       kafkapkg.TopicSimilarityRequested,
		GroupID:     groupID,
		Logger:      logger,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: -1, // LastOffset
	})

	// Create and start worker
	worker := NewSimilarityWorker(consumer, similarityService, videoEventService, logger)
	return worker.Start(ctx)
}

