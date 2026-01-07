package settings

import (
	"context"

	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
)

type Service struct {
	settingsRepo repository.SettingsRepository
	logger       *zap.Logger
}

func NewService(settingsRepo repository.SettingsRepository, logger *zap.Logger) *Service {
	return &Service{
		settingsRepo: settingsRepo,
		logger:       logger,
	}
}

func (s *Service) GetSettings(ctx context.Context) (*models.Settings, error) {
	return s.settingsRepo.GetOrCreate(ctx)
}

func (s *Service) UpdateSettings(ctx context.Context, settings *models.Settings) error {
	return s.settingsRepo.Update(ctx, settings)
}

