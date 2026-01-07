package cost

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
)

type MockCostRepository struct {
	mock.Mock
}

func (m *MockCostRepository) Create(ctx context.Context, usage *models.TokenUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *MockCostRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID) ([]*models.TokenUsage, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).([]*models.TokenUsage), args.Error(1)
}

func (m *MockCostRepository) GetSummary(ctx context.Context, startDate, endDate *time.Time) (*models.CostSummary, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CostSummary), args.Error(1)
}

func (m *MockCostRepository) GetByPeriod(ctx context.Context, period string) ([]*models.TokenUsage, error) {
	args := m.Called(ctx, period)
	return args.Get(0).([]*models.TokenUsage), args.Error(1)
}

func TestService_RecordUsage(t *testing.T) {
	mockRepo := new(MockCostRepository)
	logger := zap.NewNop()
	service := NewService(mockRepo, logger)

	ctx := context.Background()
	videoID := uuid.New()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.TokenUsage")).Return(nil)

	err := service.RecordUsage(ctx, videoID, "summarization", "gemini", "gemini-1.5-flash", 1000, 500)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetCostSummary(t *testing.T) {
	mockRepo := new(MockCostRepository)
	logger := zap.NewNop()
	service := NewService(mockRepo, logger)

	ctx := context.Background()

	expectedSummary := &models.CostSummary{
		TotalCost:  0.01,
		TotalTokens: 1000,
		ByProvider: map[string]float64{
			"gemini": 0.01,
		},
		ByOperation: map[string]float64{
			"summarization": 0.01,
		},
		ByModel: map[string]float64{
			"gemini-1.5-flash": 0.01,
		},
		Period: "month",
		VideoCount: 1,
		AverageCostPerVideo: 0.01,
	}

	mockRepo.On("GetSummary", ctx, mock.Anything, mock.Anything).Return(expectedSummary, nil)

	summary, err := service.GetCostSummary(ctx, "month")
	assert.NoError(t, err)
	assert.Equal(t, expectedSummary.TotalCost, summary.TotalCost)
	mockRepo.AssertExpectations(t)
}

func TestService_GetUsageByVideo(t *testing.T) {
	mockRepo := new(MockCostRepository)
	logger := zap.NewNop()
	service := NewService(mockRepo, logger)

	ctx := context.Background()
	videoID := uuid.New()

	expectedUsages := []*models.TokenUsage{
		{
			VideoID:      videoID,
			Operation:    "summarization",
			Provider:     "gemini",
			Model:        "gemini-1.5-flash",
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
			Cost:         0.0015,
		},
	}

	mockRepo.On("GetByVideoID", ctx, videoID).Return(expectedUsages, nil)

	usages, err := service.GetUsageByVideo(ctx, videoID)
	assert.NoError(t, err)
	assert.Len(t, usages, 1)
	mockRepo.AssertExpectations(t)
}

