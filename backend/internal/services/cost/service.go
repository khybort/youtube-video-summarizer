package cost

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/pkg/pricing"
)

type Service struct {
	costRepo repository.CostRepository
	logger   *zap.Logger
}

func NewService(costRepo repository.CostRepository, logger *zap.Logger) *Service {
	return &Service{
		costRepo: costRepo,
		logger:   logger,
	}
}

func (s *Service) RecordUsage(
	ctx context.Context,
	videoID uuid.UUID,
	operation, provider, model string,
	inputTokens, outputTokens int,
) error {
	totalTokens := inputTokens + outputTokens

	// Calculate cost
	var cost float64
	var err error

	if provider == "groq" && operation == "transcription" {
		// Groq Whisper uses per-minute pricing, not per-token
		// We'll estimate based on tokens (rough approximation)
		cost = pricing.CalculateGroqWhisperCost(float64(totalTokens) / 1000.0)
	} else {
		cost, err = pricing.CalculateCost(provider, model, inputTokens, outputTokens)
		if err != nil {
			s.logger.Warn("Failed to calculate cost", zap.Error(err))
			cost = 0
		}
	}

	usage := &models.TokenUsage{
		VideoID:      videoID,
		Operation:    operation,
		Provider:     provider,
		Model:        model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		Cost:         cost,
	}

	if err := s.costRepo.Create(ctx, usage); err != nil {
		return fmt.Errorf("failed to record token usage: %w", err)
	}

	s.logger.Info("Token usage recorded",
		zap.String("video_id", videoID.String()),
		zap.String("operation", operation),
		zap.String("provider", provider),
		zap.Int("tokens", totalTokens),
		zap.Float64("cost", cost),
	)

	return nil
}

func (s *Service) GetCostSummary(ctx context.Context, period string) (*models.CostSummary, error) {
	var startDate, endDate *time.Time
	now := time.Now()

	switch period {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		startDate = &start
		endDate = &now
	case "week":
		start := now.AddDate(0, 0, -7)
		startDate = &start
		endDate = &now
	case "month":
		start := now.AddDate(0, -1, 0)
		startDate = &start
		endDate = &now
	case "all":
		startDate = nil
		endDate = nil
	default:
		start := now.AddDate(0, -1, 0)
		startDate = &start
		endDate = &now
	}

	summary, err := s.costRepo.GetSummary(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %w", err)
	}

	summary.Period = period
	return summary, nil
}

func (s *Service) GetUsageByVideo(ctx context.Context, videoID uuid.UUID) ([]*models.TokenUsage, error) {
	return s.costRepo.GetByVideoID(ctx, videoID)
}

func (s *Service) GetUsageByPeriod(ctx context.Context, period string) ([]*models.TokenUsage, error) {
	return s.costRepo.GetByPeriod(ctx, period)
}

