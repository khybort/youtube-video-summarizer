package repository

import (
	"context"

	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
)

type SettingsRepository interface {
	Get(ctx context.Context) (*models.Settings, error)
	Update(ctx context.Context, settings *models.Settings) error
	GetOrCreate(ctx context.Context) (*models.Settings, error)
}

type settingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) Get(ctx context.Context) (*models.Settings, error) {
	var settings models.Settings
	err := r.db.WithContext(ctx).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *settingsRepository) GetOrCreate(ctx context.Context) (*models.Settings, error) {
	var settings models.Settings
	err := r.db.WithContext(ctx).First(&settings).Error
	if err == nil {
		return &settings, nil
	}
	
	// Check if it's a "not found" error
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	
	// Create default settings if not found
	defaultSettings := models.GetDefaultSettings()
	if err := r.db.WithContext(ctx).Create(defaultSettings).Error; err != nil {
		return nil, err
	}
	return defaultSettings, nil
}

func (r *settingsRepository) Update(ctx context.Context, settings *models.Settings) error {
	// Get existing settings or create if not exists
	existing, err := r.GetOrCreate(ctx)
	if err != nil {
		return err
	}
	
	// Update fields
	settings.ID = existing.ID
	return r.db.WithContext(ctx).Save(settings).Error
}

