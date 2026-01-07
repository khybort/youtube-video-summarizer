package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// This would connect to a test database
	// For now, we'll skip if no test DB is available
	connStr := "postgres://postgres:postgres@localhost:5432/youtube_analyzer_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Skip("Test database not available")
	}
	
	// AutoMigrate for tests
	db.AutoMigrate(&models.Video{})
	
	return db
}

func TestVideoRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	video := &models.Video{
		ID:          uuid.New(),
		YouTubeID:   "test_video_123",
		Title:       "Test Video",
		Description: "Test Description",
		ChannelID:   "test_channel",
		ChannelName: "Test Channel",
		Duration:    300,
		PublishedAt: time.Now(),
		Status:      "pending",
	}

	err := repo.Create(ctx, video)
	require.NoError(t, err)

	// Verify video was created
	retrieved, err := repo.GetByID(ctx, video.ID)
	require.NoError(t, err)
	assert.Equal(t, video.YouTubeID, retrieved.YouTubeID)
	assert.Equal(t, video.Title, retrieved.Title)
}

func TestVideoRepository_GetByYouTubeID(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	youtubeID := "test_video_456"
	video := &models.Video{
		ID:          uuid.New(),
		YouTubeID:   youtubeID,
		Title:       "Test Video 2",
		Description: "Test Description 2",
		ChannelID:   "test_channel",
		ChannelName: "Test Channel",
		Duration:    300,
		PublishedAt: time.Now(),
		Status:      "pending",
	}

	err := repo.Create(ctx, video)
	require.NoError(t, err)

	// Retrieve by YouTube ID
	retrieved, err := repo.GetByYouTubeID(ctx, youtubeID)
	require.NoError(t, err)
	assert.Equal(t, youtubeID, retrieved.YouTubeID)
}

func TestVideoRepository_List(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	// Create multiple videos
	for i := 0; i < 5; i++ {
		video := &models.Video{
			ID:          uuid.New(),
			YouTubeID:   "test_video_" + strconv.Itoa(i),
			Title:       "Test Video " + strconv.Itoa(i),
			Description: "Test Description",
			ChannelID:   "test_channel",
			ChannelName: "Test Channel",
			Duration:    300,
			PublishedAt: time.Now(),
			Status:      "pending",
		}
		err := repo.Create(ctx, video)
		require.NoError(t, err)
	}

	// List videos
	videos, total, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(videos), 5)
	assert.GreaterOrEqual(t, total, 5)
}

func TestVideoRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	video := &models.Video{
		ID:          uuid.New(),
		YouTubeID:   "test_video_update",
		Title:       "Original Title",
		Description: "Original Description",
		ChannelID:   "test_channel",
		ChannelName: "Test Channel",
		Duration:    300,
		PublishedAt: time.Now(),
		Status:      "pending",
	}

	err := repo.Create(ctx, video)
	require.NoError(t, err)

	// Update video
	video.Title = "Updated Title"
	video.Status = "processed"
	err = repo.Update(ctx, video)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, video.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
	assert.Equal(t, "processed", retrieved.Status)
}

func TestVideoRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	video := &models.Video{
		ID:          uuid.New(),
		YouTubeID:   "test_video_delete",
		Title:       "Test Video",
		Description: "Test Description",
		ChannelID:   "test_channel",
		ChannelName: "Test Channel",
		Duration:    300,
		PublishedAt: time.Now(),
		Status:      "pending",
	}

	err := repo.Create(ctx, video)
	require.NoError(t, err)

	// Delete video
	err = repo.Delete(ctx, video.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, video.ID)
	assert.Error(t, err)
}
