package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
)

// MockCostService implements CostService interface
type MockCostService struct {
	mock.Mock
}

func (m *MockCostService) RecordUsage(ctx context.Context, videoID uuid.UUID, operation, provider, model string, inputTokens, outputTokens int) error {
	args := m.Called(ctx, videoID, operation, provider, model, inputTokens, outputTokens)
	return args.Error(0)
}

func (m *MockCostService) GetCostSummary(ctx context.Context, period string) (*models.CostSummary, error) {
	args := m.Called(ctx, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CostSummary), args.Error(1)
}

func (m *MockCostService) GetUsageByPeriod(ctx context.Context, period string) ([]*models.TokenUsage, error) {
	args := m.Called(ctx, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TokenUsage), args.Error(1)
}

func (m *MockCostService) GetUsageByVideo(ctx context.Context, videoID uuid.UUID) ([]*models.TokenUsage, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TokenUsage), args.Error(1)
}


func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestCostHandler_GetCostSummary(t *testing.T) {
	mockService := new(MockCostService)
	logger := zap.NewNop()
	handler := &CostHandler{
		costService: mockService,
		logger:      logger,
	}

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

	mockService.On("GetCostSummary", mock.Anything, "month").Return(expectedSummary, nil)

	router := setupRouter()
	router.GET("/costs/summary", handler.GetCostSummary)

	req := httptest.NewRequest("GET", "/costs/summary?period=month", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCostHandler_GetUsage(t *testing.T) {
	mockService := new(MockCostService)
	logger := zap.NewNop()
	handler := &CostHandler{
		costService: mockService,
		logger:      logger,
	}

	expectedUsages := []*models.TokenUsage{
		{
			ID:          uuid.New(),
			VideoID:     uuid.New(),
			Operation:   "summarization",
			Provider:    "gemini",
			Model:       "gemini-1.5-flash",
			InputTokens: 1000,
			OutputTokens: 500,
			TotalTokens: 1500,
			Cost:        0.0015,
			CreatedAt:   time.Now(),
		},
	}

	mockService.On("GetUsageByPeriod", mock.Anything, "month").Return(expectedUsages, nil)

	router := setupRouter()
	router.GET("/costs/usage", handler.GetUsage)

	req := httptest.NewRequest("GET", "/costs/usage?period=month", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCostHandler_GetVideoUsage(t *testing.T) {
	mockService := new(MockCostService)
	logger := zap.NewNop()
	handler := &CostHandler{
		costService: mockService,
		logger:      logger,
	}

	videoID := uuid.New()
	expectedUsages := []*models.TokenUsage{
		{
			ID:          uuid.New(),
			VideoID:     videoID,
			Operation:   "summarization",
			Provider:    "gemini",
			Model:       "gemini-1.5-flash",
			InputTokens: 1000,
			OutputTokens: 500,
			TotalTokens: 1500,
			Cost:        0.0015,
			CreatedAt:   time.Now(),
		},
	}

	mockService.On("GetUsageByVideo", mock.Anything, videoID).Return(expectedUsages, nil)

	router := setupRouter()
	router.GET("/costs/videos/:id/usage", handler.GetVideoUsage)

	req := httptest.NewRequest("GET", "/costs/videos/"+videoID.String()+"/usage", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

