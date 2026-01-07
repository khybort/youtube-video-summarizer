package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"youtube-video-summarizer/backend/internal/services/embedding"
)

// MockVideoService implements VideoService interface
type MockVideoService struct {
	mock.Mock
}

func (m *MockVideoService) CreateFromURL(ctx context.Context, url string) (*models.Video, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Video), args.Error(1)
}

func (m *MockVideoService) GetByID(ctx context.Context, id uuid.UUID) (*models.Video, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Video), args.Error(1)
}

func (m *MockVideoService) List(ctx context.Context, limit, offset int) ([]*models.Video, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Video), args.Get(1).(int), args.Error(2)
}

func (m *MockVideoService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockVideoService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupVideoRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestVideoHandler_CreateVideo(t *testing.T) {
	mockVideoService := new(MockVideoService)
	mockTranscriptService := new(MockTranscriptService)
	mockSummaryService := new(MockSummaryService)
	mockEmbeddingService := new(MockEmbeddingService)
	mockSimilarityService := new(MockSimilarityService)
	logger := zap.NewNop()

	handler := &VideoHandler{
		videoService:      mockVideoService,
		transcriptService: mockTranscriptService,
		summaryService:    mockSummaryService,
		embeddingService:  mockEmbeddingService,
		similarityService: mockSimilarityService,
		logger:            logger,
	}

	expectedVideo := &models.Video{
		ID:        uuid.New(),
		YouTubeID: "test_video_123",
		Title:     "Test Video",
		Status:    "pending",
	}

	mockVideoService.On("CreateFromURL", mock.Anything, "test_video_123").Return(expectedVideo, nil)
	// Mock for background performAnalysis goroutine
	mockTranscriptService.On("GetOrCreateTranscript", mock.Anything, expectedVideo.ID, mock.MatchedBy(func(langCodes []string) bool {
		return len(langCodes) == 0
	})).Return(&models.Transcript{}, nil).Maybe()
	mockVideoService.On("GetByID", mock.Anything, expectedVideo.ID).Return(expectedVideo, nil).Maybe()
	mockEmbeddingService.On("GenerateVideoEmbeddings", mock.Anything, expectedVideo, mock.Anything).Return(&embedding.VideoEmbeddings{}, nil).Maybe()
	mockVideoService.On("UpdateStatus", mock.Anything, expectedVideo.ID, mock.Anything).Return(nil).Maybe()

	router := setupVideoRouter()
	router.POST("/videos", handler.CreateVideo)

	body := map[string]string{"youtube_id": "test_video_123"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/videos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockVideoService.AssertExpectations(t)
	// Give goroutine time to complete
	time.Sleep(100 * time.Millisecond)
}

func TestVideoHandler_GetVideo(t *testing.T) {
	mockVideoService := new(MockVideoService)
	mockTranscriptService := new(MockTranscriptService)
	mockSummaryService := new(MockSummaryService)
	mockEmbeddingService := new(MockEmbeddingService)
	mockSimilarityService := new(MockSimilarityService)
	logger := zap.NewNop()

	handler := &VideoHandler{
		videoService:      mockVideoService,
		transcriptService: mockTranscriptService,
		summaryService:    mockSummaryService,
		embeddingService:  mockEmbeddingService,
		similarityService: mockSimilarityService,
		logger:            logger,
	}

	videoID := uuid.New()
	expectedVideo := &models.Video{
		ID:        videoID,
		YouTubeID: "test_video_123",
		Title:     "Test Video",
		Status:    "processed",
	}

	mockVideoService.On("GetByID", mock.Anything, videoID).Return(expectedVideo, nil)

	router := setupVideoRouter()
	router.GET("/videos/:id", handler.GetVideo)

	req := httptest.NewRequest("GET", "/videos/"+videoID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockVideoService.AssertExpectations(t)
}

func TestVideoHandler_GetTranscript(t *testing.T) {
	mockVideoService := new(MockVideoService)
	mockTranscriptService := new(MockTranscriptService)
	mockSummaryService := new(MockSummaryService)
	mockEmbeddingService := new(MockEmbeddingService)
	mockSimilarityService := new(MockSimilarityService)
	logger := zap.NewNop()

	handler := &VideoHandler{
		videoService:      mockVideoService,
		transcriptService: mockTranscriptService,
		summaryService:    mockSummaryService,
		embeddingService:  mockEmbeddingService,
		similarityService: mockSimilarityService,
		logger:            logger,
	}

	videoID := uuid.New()
	expectedTranscript := &models.Transcript{
		ID:        uuid.New(),
		VideoID:   videoID,
		Content:   "Test transcript content",
		Language:  "en",
		Source:    "youtube",
	}

	// Handler first tries GetByVideoID, then GetOrCreateTranscript if not found
	mockTranscriptService.On("GetByVideoID", mock.Anything, videoID).Return(expectedTranscript, nil)

	router := setupVideoRouter()
	router.GET("/videos/:id/transcript", handler.GetTranscript)

	req := httptest.NewRequest("GET", "/videos/"+videoID.String()+"/transcript", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockTranscriptService.AssertExpectations(t)
}

func TestVideoHandler_SummarizeVideo(t *testing.T) {
	mockVideoService := new(MockVideoService)
	mockTranscriptService := new(MockTranscriptService)
	mockSummaryService := new(MockSummaryService)
	mockEmbeddingService := new(MockEmbeddingService)
	mockSimilarityService := new(MockSimilarityService)
	logger := zap.NewNop()

	handler := &VideoHandler{
		videoService:      mockVideoService,
		transcriptService: mockTranscriptService,
		summaryService:    mockSummaryService,
		embeddingService:  mockEmbeddingService,
		similarityService: mockSimilarityService,
		logger:            logger,
	}

	videoID := uuid.New()
	expectedTranscript := &models.Transcript{
		ID:        uuid.New(),
		VideoID:   videoID,
		Content:   "Test transcript content",
		Language:  "en",
		Source:    "youtube",
	}
	expectedSummary := &models.Summary{
		ID:          uuid.New(),
		VideoID:     videoID,
		Content:     "Test summary",
		SummaryType: "short",
		ModelUsed:   "gemini",
	}

	expectedVideo := &models.Video{
		ID:        videoID,
		YouTubeID: "test123",
		Title:     "Test Video",
	}
	mockVideoService.On("GetByID", mock.Anything, videoID).Return(expectedVideo, nil)
	mockTranscriptService.On("GetByVideoID", mock.Anything, videoID).Return(nil, fmt.Errorf("not found"))
	mockTranscriptService.On("GetOrCreateTranscript", mock.Anything, videoID, mock.MatchedBy(func(langCodes []string) bool {
		return len(langCodes) == 0
	})).Return(expectedTranscript, nil)
	mockSummaryService.On("GenerateSummary", mock.Anything, videoID, expectedTranscript.Content, "short", mock.Anything).Return(expectedSummary, nil)
	mockVideoService.On("UpdateStatus", mock.Anything, videoID, "completed").Return(nil)

	router := setupVideoRouter()
	router.POST("/videos/:id/summarize", handler.SummarizeVideo)

	body := map[string]string{"type": "short"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/videos/"+videoID.String()+"/summarize", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockTranscriptService.AssertExpectations(t)
	mockSummaryService.AssertExpectations(t)
	mockVideoService.AssertExpectations(t)
}

func TestVideoHandler_GetSimilarVideos(t *testing.T) {
	mockVideoService := new(MockVideoService)
	mockTranscriptService := new(MockTranscriptService)
	mockSummaryService := new(MockSummaryService)
	mockEmbeddingService := new(MockEmbeddingService)
	mockSimilarityService := new(MockSimilarityService)
	logger := zap.NewNop()

	handler := &VideoHandler{
		videoService:      mockVideoService,
		transcriptService: mockTranscriptService,
		summaryService:    mockSummaryService,
		embeddingService:  mockEmbeddingService,
		similarityService: mockSimilarityService,
		logger:            logger,
	}

	videoID := uuid.New()
	expectedVideo := &models.Video{
		ID:        videoID,
		YouTubeID: "test_video_123",
		Title:     "Test Video",
		Status:    "completed",
	}
	expectedTranscript := &models.Transcript{
		ID:        uuid.New(),
		VideoID:   videoID,
		Content:   "Test transcript content",
		Language:  "en",
		Source:    "youtube",
	}
	expectedSimilar := []models.SimilarVideo{
		{
			Video: &models.Video{
				ID:        uuid.New(),
				YouTubeID: "similar_video_1",
				Title:     "Similar Video 1",
			},
			SimilarityScore: 0.95,
			ComparisonType:  "combined",
		},
	}

	mockVideoService.On("GetByID", mock.Anything, videoID).Return(expectedVideo, nil)
	// Handler first tries GetByVideoID, then generates embeddings if transcript exists
	mockTranscriptService.On("GetByVideoID", mock.Anything, videoID).Return(expectedTranscript, nil)
	mockEmbeddingService.On("GenerateVideoEmbeddings", mock.Anything, expectedVideo, expectedTranscript.Content).Return(&embedding.VideoEmbeddings{}, nil)
	mockSimilarityService.On("FindSimilarVideos", mock.Anything, videoID, 10, 0.5).Return(expectedSimilar, nil)

	router := setupVideoRouter()
	router.GET("/videos/:id/similar", handler.GetSimilarVideos)

	req := httptest.NewRequest("GET", "/videos/"+videoID.String()+"/similar?limit=10&min_score=0.5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockVideoService.AssertExpectations(t)
	mockTranscriptService.AssertExpectations(t)
	mockEmbeddingService.AssertExpectations(t)
	mockSimilarityService.AssertExpectations(t)
}

// MockTranscriptService implements TranscriptService interface
type MockTranscriptService struct {
	mock.Mock
}

func (m *MockTranscriptService) GetOrCreateTranscript(ctx context.Context, videoID uuid.UUID, languageCode ...string) (*models.Transcript, error) {
	// Handle variadic parameter - convert to slice for mock
	var langCodes []string
	if len(languageCode) > 0 {
		langCodes = languageCode
	}
	args := m.Called(ctx, videoID, langCodes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transcript), args.Error(1)
}

func (m *MockTranscriptService) GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Transcript, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transcript), args.Error(1)
}

func (m *MockTranscriptService) DownloadAudio(youtubeID string) (string, error) {
	args := m.Called(youtubeID)
	return args.String(0), args.Error(1)
}

func (m *MockTranscriptService) ListAvailableLanguages(ctx context.Context, youtubeID string) ([]AvailableLanguage, error) {
	args := m.Called(ctx, youtubeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]AvailableLanguage), args.Error(1)
}

// MockSummaryService implements SummaryService interface
type MockSummaryService struct {
	mock.Mock
}

func (m *MockSummaryService) GenerateSummary(ctx context.Context, videoID uuid.UUID, transcript string, summaryType string, language string) (*models.Summary, error) {
	args := m.Called(ctx, videoID, transcript, summaryType, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *MockSummaryService) GenerateSummaryFromAudio(ctx context.Context, videoID uuid.UUID, audioPath string, summaryType string, language string) (*models.Summary, error) {
	args := m.Called(ctx, videoID, audioPath, summaryType, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *MockSummaryService) GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Summary, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}

// MockEmbeddingService implements EmbeddingService interface
type MockEmbeddingService struct {
	mock.Mock
}

func (m *MockEmbeddingService) GenerateVideoEmbeddings(ctx context.Context, video *models.Video, transcript string) (*embedding.VideoEmbeddings, error) {
	args := m.Called(ctx, video, transcript)
	if args.Get(0) == nil {
		if len(args) > 1 {
			return nil, args.Error(1)
		}
		return nil, nil
	}
	if len(args) > 1 {
		return args.Get(0).(*embedding.VideoEmbeddings), args.Error(1)
	}
	return args.Get(0).(*embedding.VideoEmbeddings), nil
}

// MockSimilarityService implements SimilarityService interface
type MockSimilarityService struct {
	mock.Mock
}

func (m *MockSimilarityService) FindSimilarVideos(ctx context.Context, videoID uuid.UUID, limit int, minThreshold float64) ([]models.SimilarVideo, error) {
	args := m.Called(ctx, videoID, limit, minThreshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SimilarVideo), args.Error(1)
}

