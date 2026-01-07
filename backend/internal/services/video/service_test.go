package video

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/pkg/youtube"
)

type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) Create(ctx context.Context, video *models.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Video, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Video), args.Error(1)
}

func (m *MockVideoRepository) GetByYouTubeID(ctx context.Context, youtubeID string) (*models.Video, error) {
	args := m.Called(ctx, youtubeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Video), args.Error(1)
}

func (m *MockVideoRepository) List(ctx context.Context, offset, limit int) ([]*models.Video, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*models.Video), args.Error(1)
}

func (m *MockVideoRepository) Update(ctx context.Context, video *models.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockYouTubeClient struct {
	mock.Mock
}

func (m *MockYouTubeClient) GetVideoDetails(ctx context.Context, videoID string) (*youtube.VideoDetails, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*youtube.VideoDetails), args.Error(1)
}

func TestService_Create(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	mockYouTube := new(MockYouTubeClient)
	logger := zap.NewNop()

	service := NewService(mockRepo, mockYouTube, logger)

	ctx := context.Background()
	youtubeID := "test_video_123"

	expectedDetails := &youtube.VideoDetails{
		ID:          youtubeID,
		Title:       "Test Video",
		Description: "Test Description",
		ChannelID:   "channel_123",
		ChannelName: "Test Channel",
		Duration:    300,
		PublishedAt: "2024-01-01T00:00:00Z",
	}

	mockYouTube.On("GetVideoDetails", ctx, youtubeID).Return(expectedDetails, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Video")).Return(nil)

	video, err := service.Create(ctx, youtubeID)
	assert.NoError(t, err)
	assert.NotNil(t, video)
	assert.Equal(t, youtubeID, video.YouTubeID)

	mockRepo.AssertExpectations(t)
	mockYouTube.AssertExpectations(t)
}

func TestService_GetByID(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	mockYouTube := new(MockYouTubeClient)
	logger := zap.NewNop()

	service := NewService(mockRepo, mockYouTube, logger)

	ctx := context.Background()
	videoID := uuid.New()

	expectedVideo := &models.Video{
		ID:        videoID,
		YouTubeID: "test_video_123",
		Title:     "Test Video",
		Status:    "processed",
	}

	mockRepo.On("GetByID", ctx, videoID).Return(expectedVideo, nil)

	video, err := service.GetByID(ctx, videoID)
	assert.NoError(t, err)
	assert.Equal(t, expectedVideo, video)

	mockRepo.AssertExpectations(t)
}

func TestService_List(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	mockYouTube := new(MockYouTubeClient)
	logger := zap.NewNop()

	service := NewService(mockRepo, mockYouTube, logger)

	ctx := context.Background()

	expectedVideos := []*models.Video{
		{ID: uuid.New(), YouTubeID: "video1", Title: "Video 1"},
		{ID: uuid.New(), YouTubeID: "video2", Title: "Video 2"},
	}

	mockRepo.On("List", ctx, 0, 10).Return(expectedVideos, nil)

	videos, err := service.List(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Len(t, videos, 2)

	mockRepo.AssertExpectations(t)
}

