package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
)

type TranscriptRepository interface {
	Create(ctx context.Context, transcript *models.Transcript) error
	GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Transcript, error)
	Update(ctx context.Context, transcript *models.Transcript) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type transcriptRepository struct {
	db *gorm.DB
}

func NewTranscriptRepository(db *gorm.DB) TranscriptRepository {
	return &transcriptRepository{db: db}
}

func (r *transcriptRepository) Create(ctx context.Context, transcript *models.Transcript) error {
	if transcript.ID == uuid.Nil {
		transcript.ID = uuid.New()
	}
	// GORM will automatically handle JSONB serialization for []TranscriptSegment
	return r.db.WithContext(ctx).Create(transcript).Error
}

func (r *transcriptRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Transcript, error) {
	var transcript models.Transcript
	err := r.db.WithContext(ctx).
		Where("video_id = ?", videoID).
		Order("created_at DESC").
		First(&transcript).Error
	if err != nil {
		return nil, err
	}
	return &transcript, nil
}

func (r *transcriptRepository) Update(ctx context.Context, transcript *models.Transcript) error {
	return r.db.WithContext(ctx).Save(transcript).Error
}

func (r *transcriptRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Transcript{}, id).Error
}
