package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
)

type SummaryRepository interface {
	Create(ctx context.Context, summary *models.Summary) error
	GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Summary, error)
}

type summaryRepository struct {
	db *gorm.DB
}

func NewSummaryRepository(db *gorm.DB) SummaryRepository {
	return &summaryRepository{db: db}
}

func (r *summaryRepository) Create(ctx context.Context, summary *models.Summary) error {
	if summary.ID == uuid.Nil {
		summary.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(summary).Error
}

func (r *summaryRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Summary, error) {
	var summary models.Summary
	err := r.db.WithContext(ctx).
		Where("video_id = ?", videoID).
		Order("created_at DESC").
		First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}
