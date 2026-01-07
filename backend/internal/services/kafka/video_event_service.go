package kafka

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/pkg/kafka"
)

// VideoEventService handles video-related Kafka events
type VideoEventService struct {
	producer *kafka.Producer
	logger   *zap.Logger
}

// NewVideoEventService creates a new VideoEventService
func NewVideoEventService(producer *kafka.Producer, logger *zap.Logger) *VideoEventService {
	return &VideoEventService{
		producer: producer,
		logger:   logger,
	}
}

// PublishVideoCreated publishes a video.created event
func (s *VideoEventService) PublishVideoCreated(ctx context.Context, video *models.Video) error {
	event := kafka.NewVideoCreatedEvent(
		video.ID.String(),
		video.YouTubeID,
		video.Title,
		video.ChannelID,
		video.ChannelName,
		video.Duration,
		video.PublishedAt,
		video.ThumbnailURL,
	)

	if err := s.producer.Publish(ctx, kafka.TopicVideoCreated, video.ID.String(), event); err != nil {
		s.logger.Error("Failed to publish video.created event",
			zap.String("video_id", video.ID.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Published video.created event",
		zap.String("video_id", video.ID.String()),
		zap.String("youtube_id", video.YouTubeID),
	)

	return nil
}

// PublishTranscriptRequested publishes a transcript.requested event
func (s *VideoEventService) PublishTranscriptRequested(ctx context.Context, videoID uuid.UUID, youtubeID string, priority int) error {
	event := kafka.NewTranscriptRequestedEvent(videoID.String(), youtubeID, priority)

	if err := s.producer.Publish(ctx, kafka.TopicTranscriptRequested, videoID.String(), event); err != nil {
		s.logger.Error("Failed to publish transcript.requested event",
			zap.String("video_id", videoID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// PublishEmbeddingRequested publishes an embedding.requested event
func (s *VideoEventService) PublishEmbeddingRequested(ctx context.Context, videoID uuid.UUID, youtubeID, transcriptContent string, priority int) error {
	event := kafka.NewEmbeddingRequestedEvent(videoID.String(), youtubeID, transcriptContent, priority)

	if err := s.producer.Publish(ctx, kafka.TopicEmbeddingRequested, videoID.String(), event); err != nil {
		s.logger.Error("Failed to publish embedding.requested event",
			zap.String("video_id", videoID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// PublishSimilarityRequested publishes a similarity.requested event
func (s *VideoEventService) PublishSimilarityRequested(ctx context.Context, videoID uuid.UUID, youtubeID, targetVideoID string, priority int) error {
	event := kafka.NewSimilarityRequestedEvent(videoID.String(), youtubeID, targetVideoID, priority)

	if err := s.producer.Publish(ctx, kafka.TopicSimilarityRequested, videoID.String(), event); err != nil {
		s.logger.Error("Failed to publish similarity.requested event",
			zap.String("video_id", videoID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// PublishAnalysisCompleted publishes an analysis.completed event
func (s *VideoEventService) PublishAnalysisCompleted(ctx context.Context, videoID uuid.UUID, youtubeID string, hasTranscript, hasSummary, hasEmbedding bool, duration int) error {
	event := kafka.NewAnalysisCompletedEvent(videoID.String(), youtubeID, hasTranscript, hasSummary, hasEmbedding, duration)

	if err := s.producer.Publish(ctx, kafka.TopicAnalysisCompleted, videoID.String(), event); err != nil {
		s.logger.Error("Failed to publish analysis.completed event",
			zap.String("video_id", videoID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// PublishAnalysisFailed publishes an analysis.failed event
func (s *VideoEventService) PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error {
	event := kafka.NewAnalysisFailedEvent(videoID.String(), youtubeID, stage, errorMsg, retryable)

	if err := s.producer.Publish(ctx, kafka.TopicAnalysisFailed, videoID.String(), event); err != nil {
		s.logger.Error("Failed to publish analysis.failed event",
			zap.String("video_id", videoID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

