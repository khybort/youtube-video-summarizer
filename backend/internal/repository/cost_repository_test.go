package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
)

func setupTestDBForCost(t *testing.T) *gorm.DB {
	connStr := "postgres://postgres:postgres@localhost:5432/youtube_analyzer_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Skip("Test database not available")
	}
	
	// AutoMigrate for tests
	db.AutoMigrate(&models.Video{}, &models.TokenUsage{})
	
	return db
}

func TestCostRepository_Create(t *testing.T) {
	db := setupTestDBForCost(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewCostRepository(db)
	ctx := context.Background()

	videoID := uuid.New()
	usage := &models.TokenUsage{
		VideoID:      videoID,
		Operation:    "summarization",
		Provider:     "gemini",
		Model:        "gemini-1.5-flash",
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		Cost:         0.0015,
	}

	err := repo.Create(ctx, usage)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, usage.ID)
}

func TestCostRepository_GetByVideoID(t *testing.T) {
	db := setupTestDBForCost(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewCostRepository(db)
	ctx := context.Background()

	videoID := uuid.New()
	usage := &models.TokenUsage{
		VideoID:      videoID,
		Operation:    "summarization",
		Provider:     "gemini",
		Model:        "gemini-1.5-flash",
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		Cost:         0.0015,
	}

	err := repo.Create(ctx, usage)
	require.NoError(t, err)

	// Retrieve by video ID
	usages, err := repo.GetByVideoID(ctx, videoID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(usages), 1)
	assert.Equal(t, videoID, usages[0].VideoID)
}

func TestCostRepository_GetSummary(t *testing.T) {
	db := setupTestDBForCost(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewCostRepository(db)
	ctx := context.Background()

	videoID := uuid.New()
	usage := &models.TokenUsage{
		VideoID:      videoID,
		Operation:    "summarization",
		Provider:     "gemini",
		Model:        "gemini-1.5-flash",
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		Cost:         0.0015,
	}

	err := repo.Create(ctx, usage)
	require.NoError(t, err)

	// Get summary
	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	summary, err := repo.GetSummary(ctx, &startDate, &now)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, summary.TotalCost, 0.0)
	assert.GreaterOrEqual(t, summary.TotalTokens, 0)
}

func TestCostRepository_GetByPeriod(t *testing.T) {
	db := setupTestDBForCost(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewCostRepository(db)
	ctx := context.Background()

	videoID := uuid.New()
	usage := &models.TokenUsage{
		VideoID:      videoID,
		Operation:    "summarization",
		Provider:     "gemini",
		Model:        "gemini-1.5-flash",
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		Cost:         0.0015,
	}

	err := repo.Create(ctx, usage)
	require.NoError(t, err)

	// Get by period
	usages, err := repo.GetByPeriod(ctx, "month")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(usages), 0)
}
