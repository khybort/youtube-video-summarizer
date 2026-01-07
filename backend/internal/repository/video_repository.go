package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
)

type VideoRepository interface {
	Create(ctx context.Context, video *models.Video) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Video, error)
	GetByYouTubeID(ctx context.Context, youtubeID string) (*models.Video, error)
	List(ctx context.Context, limit, offset int) ([]*models.Video, int, error)
	Update(ctx context.Context, video *models.Video) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type videoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepository{db: db}
}

func (r *videoRepository) Create(ctx context.Context, video *models.Video) error {
	if video.ID == uuid.Nil {
		video.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *videoRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Video, error) {
	var video models.Video
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *videoRepository) GetByYouTubeID(ctx context.Context, youtubeID string) (*models.Video, error) {
	var video models.Video
	err := r.db.WithContext(ctx).Where("youtube_id = ?", youtubeID).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *videoRepository) List(ctx context.Context, limit, offset int) ([]*models.Video, int, error) {
	var videos []*models.Video
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Video{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get videos
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&videos).Error

	if err != nil {
		return nil, 0, err
	}

	return videos, int(total), nil
}

func (r *videoRepository) Update(ctx context.Context, video *models.Video) error {
	return r.db.WithContext(ctx).Save(video).Error
}

func (r *videoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Video{}, id).Error
}

func (r *videoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Video{}).
		Where("id = ?", id).
		Update("status", status).Error
}
