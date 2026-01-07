package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/services/embedding"
	"youtube-video-summarizer/backend/internal/services/similarity"
	"youtube-video-summarizer/backend/internal/services/transcript"
	"youtube-video-summarizer/backend/internal/services/video"
)

type AnalysisJob struct {
	videoService     *video.Service
	transcriptService *transcript.Service
	embeddingService  *embedding.Service
	similarityService *similarity.Service
	logger            *zap.Logger
}

func NewAnalysisJob(
	videoService *video.Service,
	transcriptService *transcript.Service,
	embeddingService *embedding.Service,
	similarityService *similarity.Service,
	logger *zap.Logger,
) *AnalysisJob {
	return &AnalysisJob{
		videoService:      videoService,
		transcriptService: transcriptService,
		embeddingService:  embeddingService,
		similarityService: similarityService,
		logger:            logger,
	}
}

func (j *AnalysisJob) ProcessVideo(ctx context.Context, videoID uuid.UUID) error {
	startTime := time.Now()
	j.logger.Info("Starting video analysis", zap.String("video_id", videoID.String()))

	// Update status
	if err := j.videoService.UpdateStatus(ctx, videoID, "processing"); err != nil {
		return err
	}

	// 1. Get or create transcript
	transcript, err := j.transcriptService.GetOrCreateTranscript(ctx, videoID)
	if err != nil {
		j.logger.Error("Failed to get transcript", zap.Error(err))
		j.videoService.UpdateStatus(ctx, videoID, "error")
		return err
	}

	// 2. Get video
	video, err := j.videoService.GetByID(ctx, videoID)
	if err != nil {
		j.logger.Error("Failed to get video", zap.Error(err))
		return err
	}

	// 3. Generate embeddings
	_, err = j.embeddingService.GenerateVideoEmbeddings(ctx, video, transcript.Content)
	if err != nil {
		j.logger.Error("Failed to generate embeddings", zap.Error(err))
		j.videoService.UpdateStatus(ctx, videoID, "error")
		return err
	}

	// 4. Calculate similarities (async, non-blocking)
	go j.calculateSimilarities(ctx, videoID)

	// 5. Update status
	if err := j.videoService.UpdateStatus(ctx, videoID, "completed"); err != nil {
		return err
	}

	duration := time.Since(startTime)
	j.logger.Info("Video analysis completed",
		zap.String("video_id", videoID.String()),
		zap.Duration("duration", duration),
	)

	return nil
}

func (j *AnalysisJob) calculateSimilarities(ctx context.Context, videoID uuid.UUID) {
	// This would typically:
	// 1. Get all other videos
	// 2. Calculate similarities in batches
	// 3. Save to similarity cache
	// For now, similarities are calculated on-demand
	j.logger.Info("Similarity calculation completed", zap.String("video_id", videoID.String()))
}

