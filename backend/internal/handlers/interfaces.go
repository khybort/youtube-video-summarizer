package handlers

import (
	"context"

	"github.com/google/uuid"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/services/embedding"
	"youtube-video-summarizer/backend/internal/services/transcript"
)

// Service interfaces for dependency injection and testing

type VideoService interface {
	CreateFromURL(ctx context.Context, url string) (*models.Video, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Video, error)
	List(ctx context.Context, limit, offset int) ([]*models.Video, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AvailableLanguage is an alias for transcript.AvailableLanguage
type AvailableLanguage = transcript.AvailableLanguage

type TranscriptService interface {
	GetOrCreateTranscript(ctx context.Context, videoID uuid.UUID, languageCode ...string) (*models.Transcript, error)
	GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Transcript, error)
	DownloadAudio(youtubeID string) (string, error)
	ListAvailableLanguages(ctx context.Context, youtubeID string) ([]transcript.AvailableLanguage, error)
}

type SummaryService interface {
	GenerateSummary(ctx context.Context, videoID uuid.UUID, transcript string, summaryType string, language string) (*models.Summary, error)
	GenerateSummaryFromAudio(ctx context.Context, videoID uuid.UUID, audioPath string, summaryType string, language string) (*models.Summary, error)
	GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Summary, error)
}

type EmbeddingService interface {
	GenerateVideoEmbeddings(ctx context.Context, video *models.Video, transcript string) (*embedding.VideoEmbeddings, error)
}

type CostService interface {
	RecordUsage(ctx context.Context, videoID uuid.UUID, operation, provider, model string, inputTokens, outputTokens int) error
	GetCostSummary(ctx context.Context, period string) (*models.CostSummary, error)
	GetUsageByPeriod(ctx context.Context, period string) ([]*models.TokenUsage, error)
	GetUsageByVideo(ctx context.Context, videoID uuid.UUID) ([]*models.TokenUsage, error)
}

// VideoCostBreakdown is returned by CostService
type VideoCostBreakdown struct {
	TotalCost        float64
	TotalInputTokens  int
	TotalOutputTokens int
	Breakdown         []CostBreakdownItem
}

type CostBreakdownItem struct {
	Operation    string
	Provider     string
	Model        string
	Cost         float64
	InputTokens  int
	OutputTokens int
}

type SimilarityService interface {
	FindSimilarVideos(ctx context.Context, videoID uuid.UUID, limit int, minThreshold float64) ([]models.SimilarVideo, error)
}

