package video

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/pkg/youtube"
)

type Service struct {
	videoRepo     repository.VideoRepository
	youtubeClient *youtube.Client
	logger        *zap.Logger
}

func NewService(
	videoRepo repository.VideoRepository,
	youtubeClient *youtube.Client,
	logger *zap.Logger,
) *Service {
	return &Service{
		videoRepo:     videoRepo,
		youtubeClient: youtubeClient,
		logger:        logger,
	}
}

func (s *Service) CreateFromURL(ctx context.Context, videoURL string) (*models.Video, error) {
	// Extract video ID
	videoID, err := s.youtubeClient.ExtractVideoID(videoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid YouTube URL: %w", err)
	}

	// Check if video already exists
	existing, err := s.videoRepo.GetByYouTubeID(ctx, videoID)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Fetch video info from YouTube
	info, err := s.youtubeClient.GetVideoInfo(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video info: %w", err)
	}

	// Create video model
	video := &models.Video{
		YouTubeID:    info.ID,
		Title:        info.Title,
		Description:  info.Description,
		ChannelID:     info.ChannelID,
		ChannelName:  info.ChannelName,
		Duration:     info.Duration,
		ViewCount:    info.ViewCount,
		LikeCount:    info.LikeCount,
		PublishedAt:  info.PublishedAt,
		ThumbnailURL: info.ThumbnailURL,
		Tags:         info.Tags,
		Category:     info.Category,
		Status:       "pending",
		HasTranscript: false,
		HasSummary:   false,
	}

	// Save to database
	if err := s.videoRepo.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("failed to save video: %w", err)
	}

	s.logger.Info("Video created", zap.String("video_id", video.ID.String()), zap.String("youtube_id", videoID))

	return video, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Video, error) {
	return s.videoRepo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*models.Video, int, error) {
	return s.videoRepo.List(ctx, limit, offset)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.videoRepo.Delete(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.videoRepo.UpdateStatus(ctx, id, status)
}

