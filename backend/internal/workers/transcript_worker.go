package workers

import (
	"context"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/internal/services/cost"
	"youtube-video-summarizer/backend/internal/services/provider"
	"youtube-video-summarizer/backend/internal/services/transcript"
	kafkapkg "youtube-video-summarizer/backend/pkg/kafka"
	"youtube-video-summarizer/backend/pkg/youtube"
)

// TranscriptWorker processes transcript generation requests from Kafka
type TranscriptWorker struct {
	consumer        *kafkapkg.Consumer
	transcriptService *transcript.Service
	videoEventService interface {
		PublishEmbeddingRequested(ctx context.Context, videoID uuid.UUID, youtubeID, transcriptContent string, priority int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	}
	logger *zap.Logger
}

// NewTranscriptWorker creates a new TranscriptWorker
func NewTranscriptWorker(
	consumer *kafkapkg.Consumer,
	transcriptService *transcript.Service,
	videoEventService interface {
		PublishEmbeddingRequested(ctx context.Context, videoID uuid.UUID, youtubeID, transcriptContent string, priority int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	},
	logger *zap.Logger,
) *TranscriptWorker {
	return &TranscriptWorker{
		consumer:          consumer,
		transcriptService: transcriptService,
		videoEventService: videoEventService,
		logger:            logger,
	}
}

// Start starts the transcript worker
func (w *TranscriptWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, w.handleMessage)
}

// handleMessage processes a transcript request message
func (w *TranscriptWorker) handleMessage(ctx context.Context, message kafkago.Message) error {
	// Parse event
	var event kafkapkg.TranscriptRequestedEvent
	if err := kafkapkg.UnmarshalEvent(message, &event); err != nil {
		w.logger.Error("Failed to unmarshal transcript.requested event", zap.Error(err))
		return err
	}

	videoID, err := uuid.Parse(event.VideoID)
	if err != nil {
		w.logger.Error("Invalid video ID in event", zap.String("video_id", event.VideoID), zap.Error(err))
		return err
	}

	w.logger.Info("Processing transcript request",
		zap.String("video_id", event.VideoID),
		zap.String("youtube_id", event.YouTubeID),
		zap.Int("priority", event.Priority),
	)

	// Get or create transcript
	transcript, err := w.transcriptService.GetOrCreateTranscript(ctx, videoID)
	if err != nil {
		w.logger.Error("Failed to generate transcript",
			zap.String("video_id", event.VideoID),
			zap.Error(err),
		)
		
		// Publish failure event
		_ = w.videoEventService.PublishAnalysisFailed(
			ctx,
			videoID,
			event.YouTubeID,
			"transcript",
			err.Error(),
			true, // Retryable
		)
		return err
	}

	w.logger.Info("Transcript generated successfully",
		zap.String("video_id", event.VideoID),
		zap.String("source", transcript.Source),
		zap.Int("length", len(transcript.Content)),
	)

	// Publish embedding request event
	if err := w.videoEventService.PublishEmbeddingRequested(
		ctx,
		videoID,
		event.YouTubeID,
		transcript.Content,
		event.Priority,
	); err != nil {
		w.logger.Error("Failed to publish embedding.requested event",
			zap.String("video_id", event.VideoID),
			zap.Error(err),
		)
		// Don't fail the message - transcript was generated successfully
	}

	return nil
}

// StartTranscriptWorker starts a transcript worker with the given configuration
func StartTranscriptWorker(
	ctx context.Context,
	brokers []string,
	groupID string,
	transcriptRepo repository.TranscriptRepository,
	videoRepo repository.VideoRepository,
	providerFactory *provider.ProviderFactory,
	youtubeClient *youtube.Client,
	costService *cost.Service,
	videoEventService interface {
		PublishEmbeddingRequested(ctx context.Context, videoID uuid.UUID, youtubeID, transcriptContent string, priority int) error
		PublishAnalysisFailed(ctx context.Context, videoID uuid.UUID, youtubeID, stage, errorMsg string, retryable bool) error
	},
	logger *zap.Logger,
) error {
	// Create transcript service
	transcriptService := transcript.NewService(
		transcriptRepo,
		videoRepo,
		providerFactory,
		youtubeClient,
		costService,
		logger,
	)

	// Create consumer
	consumer := kafkapkg.NewConsumer(kafkapkg.ConsumerConfig{
		Brokers:     brokers,
		Topic:       kafkapkg.TopicTranscriptRequested,
		GroupID:     groupID,
		Logger:      logger,
		MinBytes:    1, // Minimum bytes - reduce to 1 to avoid waiting for more data
		MaxBytes:    10e6, // 10MB
		MaxWait:     10 * time.Second, // Increased to reduce CPU usage when no messages
		StartOffset: -1, // LastOffset
	})

	// Create and start worker
	worker := NewTranscriptWorker(consumer, transcriptService, videoEventService, logger)
	return worker.Start(ctx)
}

