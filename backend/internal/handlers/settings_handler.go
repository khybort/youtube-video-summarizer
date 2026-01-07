package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/config"
	settingsservice "youtube-video-summarizer/backend/internal/services/settings"
)

func RegisterSettingsRoutes(router *gin.RouterGroup, settingsService *settingsservice.Service, cfg *config.Config, logger *zap.Logger) {
	handler := &SettingsHandler{
		settingsService: settingsService,
		config:          cfg,
		logger:          logger,
	}
	
	settings := router.Group("/settings")
	{
		settings.GET("", handler.GetSettings)
		settings.PUT("", handler.UpdateSettings)
		settings.GET("/health/local-whisper", handler.CheckLocalWhisperHealth)
	}
}

type SettingsHandler struct {
	settingsService *settingsservice.Service
	config          *config.Config
	logger          *zap.Logger
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsService.GetSettings(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	// Do not return API keys; only return whether a key is configured in DB.
	c.JSON(http.StatusOK, gin.H{
		"id":                      settings.ID,
		"transcript_provider":     settings.TranscriptProvider,
		"summary_provider":        settings.SummaryProvider,
		"embedding_provider":      settings.EmbeddingProvider,
		"audio_analysis_provider": settings.AudioAnalysisProvider,
		"ollama_model":            settings.OllamaModel,
		"whisper_model":           settings.WhisperModel,
		"gemini_model":            settings.GeminiModel,
		"ollama_url":              settings.OllamaURL,
		"local_whisper_url":       settings.LocalWhisperURL,
		"summary_language":        settings.SummaryLanguage,
		"has_gemini_api_key":      settings.GeminiAPIKey != "",
		"has_groq_api_key":        settings.GroqAPIKey != "",
		"has_huggingface_api_key": settings.HuggingFaceAPIKey != "",
		"created_at":              settings.CreatedAt,
		"updated_at":              settings.UpdatedAt,
	})
}

type SettingsUpdateRequest struct {
	TranscriptProvider     *string `json:"transcript_provider,omitempty"`
	SummaryProvider        *string `json:"summary_provider,omitempty"`
	EmbeddingProvider      *string `json:"embedding_provider,omitempty"`
	AudioAnalysisProvider  *string `json:"audio_analysis_provider,omitempty"`
	OllamaModel            *string `json:"ollama_model,omitempty"`
	WhisperModel           *string `json:"whisper_model,omitempty"`
	GeminiModel            *string `json:"gemini_model,omitempty"`
	OllamaURL              *string `json:"ollama_url,omitempty"`
	LocalWhisperURL        *string `json:"local_whisper_url,omitempty"`
	SummaryLanguage        *string `json:"summary_language,omitempty"`
	GeminiAPIKey           *string `json:"gemini_api_key,omitempty"`
	GroqAPIKey             *string `json:"groq_api_key,omitempty"`
	HuggingFaceAPIKey      *string `json:"huggingface_api_key,omitempty"`
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var req SettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If user is trying to select local whisper, check if it's available
	if req.TranscriptProvider != nil && *req.TranscriptProvider == "local" {
		if !h.isLocalWhisperAvailable(c.Request.Context()) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Local Whisper service is not available. Please ensure the service is running or select a different provider.",
				"code":  "LOCAL_WHISPER_UNAVAILABLE",
			})
			return
		}
	}

	// Merge update into existing settings so omitted fields don't get zeroed.
	existing, err := h.settingsService.GetSettings(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get settings for update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	// Regular fields
	if req.TranscriptProvider != nil {
		existing.TranscriptProvider = *req.TranscriptProvider
	}
	if req.SummaryProvider != nil {
		existing.SummaryProvider = *req.SummaryProvider
	}
	if req.EmbeddingProvider != nil {
		existing.EmbeddingProvider = *req.EmbeddingProvider
	}
	if req.AudioAnalysisProvider != nil {
		existing.AudioAnalysisProvider = *req.AudioAnalysisProvider
	}
	if req.OllamaModel != nil {
		existing.OllamaModel = *req.OllamaModel
	}
	if req.WhisperModel != nil {
		existing.WhisperModel = *req.WhisperModel
	}
	if req.GeminiModel != nil {
		existing.GeminiModel = *req.GeminiModel
	}
	if req.OllamaURL != nil {
		existing.OllamaURL = *req.OllamaURL
	}
	if req.LocalWhisperURL != nil {
		existing.LocalWhisperURL = *req.LocalWhisperURL
	}
	if req.SummaryLanguage != nil {
		existing.SummaryLanguage = *req.SummaryLanguage
	}

	// LLM API keys (optional): only update if provided; allow clearing by sending "".
	if req.GeminiAPIKey != nil {
		existing.GeminiAPIKey = strings.TrimSpace(*req.GeminiAPIKey)
	}

	// API keys (optional): only update if provided; allow clearing by sending "".
	if req.GroqAPIKey != nil {
		existing.GroqAPIKey = strings.TrimSpace(*req.GroqAPIKey)
	}
	if req.HuggingFaceAPIKey != nil {
		existing.HuggingFaceAPIKey = strings.TrimSpace(*req.HuggingFaceAPIKey)
	}

	if err := h.settingsService.UpdateSettings(c.Request.Context(), existing); err != nil {
		h.logger.Error("Failed to update settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

func (h *SettingsHandler) CheckLocalWhisperHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	isAvailable := h.isLocalWhisperAvailable(ctx)
	
	c.JSON(http.StatusOK, gin.H{
		"available": isAvailable,
		"url":       h.getLocalWhisperURL(),
	})
}

func (h *SettingsHandler) isLocalWhisperAvailable(ctx context.Context) bool {
	url := h.getLocalWhisperURL()
	if url == "" {
		return false
	}

	// Try to connect to local whisper service
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Try a simple GET request to check if service is up
	// Most whisper services have a health endpoint or at least respond to GET
	req, err := http.NewRequestWithContext(ctx, "GET", url+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		// If connection fails, try with Docker service name (whisper:8001)
		// This works when backend is running in Docker
		dockerURL := "http://whisper:8001"
		if url != dockerURL {
			dockerReq, dockerErr := http.NewRequestWithContext(ctx, "GET", dockerURL+"/health", nil)
			if dockerErr == nil {
				dockerResp, dockerErr := client.Do(dockerReq)
				if dockerErr == nil {
					defer dockerResp.Body.Close()
					return dockerResp.StatusCode < 500
				}
			}
		}
		return false
	}
	defer resp.Body.Close()

	// If we get any response (even 404), the service is running
	return resp.StatusCode < 500
}

func (h *SettingsHandler) getLocalWhisperURL() string {
	if h.config != nil && h.config.Whisper.LocalWhisperURL != "" {
		return h.config.Whisper.LocalWhisperURL
	}
	return "http://localhost:8001"
}

