package similarity

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
)

type MockSimilarityRepository struct {
	mock.Mock
}

func (m *MockSimilarityRepository) Save(ctx context.Context, result *models.SimilarityResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockSimilarityRepository) GetByVideoPair(ctx context.Context, videoID1, videoID2 uuid.UUID) (*models.SimilarityResult, error) {
	args := m.Called(ctx, videoID1, videoID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SimilarityResult), args.Error(1)
}

func (m *MockSimilarityRepository) GetSimilarVideos(ctx context.Context, videoID uuid.UUID, limit int, minScore float64) ([]models.SimilarVideo, error) {
	args := m.Called(ctx, videoID, limit, minScore)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SimilarVideo), args.Error(1)
}

type MockEmbeddingRepository struct {
	mock.Mock
}

func (m *MockEmbeddingRepository) Save(ctx context.Context, embedding *models.VideoEmbedding) error {
	args := m.Called(ctx, embedding)
	return args.Error(0)
}

func (m *MockEmbeddingRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID, embeddingType string) (*models.VideoEmbedding, error) {
	args := m.Called(ctx, videoID, embeddingType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VideoEmbedding), args.Error(1)
}

func (m *MockEmbeddingRepository) FindSimilar(ctx context.Context, embedding []float32, embeddingType string, limit int, excludeVideoID uuid.UUID) ([]repository.SimilarEmbedding, error) {
	args := m.Called(ctx, embedding, embeddingType, limit, excludeVideoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.SimilarEmbedding), args.Error(1)
}

func (m *MockEmbeddingRepository) CountVideosWithEmbeddings(ctx context.Context, embeddingType string, excludeVideoID uuid.UUID) (int, error) {
	args := m.Called(ctx, embeddingType, excludeVideoID)
	return args.Int(0), args.Error(1)
}

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

func (m *MockVideoRepository) Update(ctx context.Context, video *models.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockVideoRepository) List(ctx context.Context, limit, offset int) ([]*models.Video, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Video), args.Get(1).(int), args.Error(2)
}

func (m *MockVideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestService_CalculateSimilarity(t *testing.T) {
	mockSimilarityRepo := new(MockSimilarityRepository)
	mockEmbeddingRepo := new(MockEmbeddingRepository)
	mockVideoRepo := new(MockVideoRepository)
	logger := zap.NewNop()

	service := NewService(mockSimilarityRepo, mockEmbeddingRepo, mockVideoRepo, nil, logger)

	ctx := context.Background()
	videoID1 := uuid.New()
	videoID2 := uuid.New()

	embedding1 := &models.VideoEmbedding{
		VideoID:       videoID1,
		EmbeddingType: "combined",
		Embedding:     models.Vector{Data: []float32{0.1, 0.2, 0.3}},
	}

	embedding2 := &models.VideoEmbedding{
		VideoID:       videoID2,
		EmbeddingType: "combined",
		Embedding:     models.Vector{Data: []float32{0.1, 0.2, 0.3}},
	}

	mockEmbeddingRepo.On("GetByVideoID", ctx, videoID1, "combined").Return(embedding1, nil).Once()
	mockEmbeddingRepo.On("GetByVideoID", ctx, videoID2, "combined").Return(embedding2, nil).Once()
	// getSimilarityByType returns 0 if embedding not found, so we can return nil, error
	mockEmbeddingRepo.On("GetByVideoID", ctx, videoID1, "title").Return(nil, fmt.Errorf("not found")).Maybe()
	mockEmbeddingRepo.On("GetByVideoID", ctx, videoID2, "title").Return(nil, fmt.Errorf("not found")).Maybe()
	mockEmbeddingRepo.On("GetByVideoID", ctx, videoID1, "description").Return(nil, fmt.Errorf("not found")).Maybe()
	mockEmbeddingRepo.On("GetByVideoID", ctx, videoID2, "description").Return(nil, fmt.Errorf("not found")).Maybe()
	mockEmbeddingRepo.On("GetByVideoID", ctx, videoID1, "transcript").Return(nil, fmt.Errorf("not found")).Maybe()
	mockEmbeddingRepo.On("GetByVideoID", ctx, videoID2, "transcript").Return(nil, fmt.Errorf("not found")).Maybe()
	mockSimilarityRepo.On("Save", ctx, mock.AnythingOfType("*models.SimilarityResult")).Return(nil).Maybe()

	similarity, err := service.CalculateSimilarity(ctx, videoID1, videoID2)
	assert.NoError(t, err)
	assert.NotNil(t, similarity)
	assert.Greater(t, similarity.CombinedSimilarity, 0.0)

	mockSimilarityRepo.AssertExpectations(t)
	mockEmbeddingRepo.AssertExpectations(t)
}

func TestService_FindSimilarVideos(t *testing.T) {
	mockSimilarityRepo := new(MockSimilarityRepository)
	mockEmbeddingRepo := new(MockEmbeddingRepository)
	mockVideoRepo := new(MockVideoRepository)
	logger := zap.NewNop()

	// Service requires YouTube client, so we expect an error when it's nil
	service := NewService(mockSimilarityRepo, mockEmbeddingRepo, mockVideoRepo, nil, logger)

	ctx := context.Background()
	videoID := uuid.New()

	// Mock GetByID call (service calls this first)
	expectedVideo := &models.Video{
		ID:        videoID,
		YouTubeID: "test123",
	}
	mockVideoRepo.On("GetByID", ctx, videoID).Return(expectedVideo, nil)

	// Since YouTube client is nil, we expect an error
	_, err := service.FindSimilarVideos(ctx, videoID, 10, 0.5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YouTube client not configured")
	
	mockVideoRepo.AssertExpectations(t)
}

