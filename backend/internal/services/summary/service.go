package summary

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/config"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/internal/services/cost"
	settingsservice "youtube-video-summarizer/backend/internal/services/settings"
	"youtube-video-summarizer/backend/internal/services/provider"
	"youtube-video-summarizer/backend/pkg/llm"
	"youtube-video-summarizer/backend/pkg/prompts"
	"youtube-video-summarizer/backend/pkg/whisper"
)

type Service struct {
	summaryRepo     repository.SummaryRepository
	providerFactory *provider.ProviderFactory
	costService     *cost.Service
	settingsService *settingsservice.Service
	logger          *zap.Logger
	tempDir         string
	config          *config.Config
}

func NewService(
	summaryRepo repository.SummaryRepository,
	providerFactory *provider.ProviderFactory,
	costService *cost.Service,
	settingsService *settingsservice.Service,
	logger *zap.Logger,
	cfg *config.Config,
) *Service {
	tempDir := os.TempDir()
	return &Service{
		summaryRepo:     summaryRepo,
		providerFactory: providerFactory,
		costService:     costService,
		settingsService: settingsService,
		logger:          logger,
		tempDir:         tempDir,
		config:          cfg,
	}
}

func (s *Service) GetByVideoID(ctx context.Context, videoID uuid.UUID) (*models.Summary, error) {
	return s.summaryRepo.GetByVideoID(ctx, videoID)
}

// GenerateSummaryFromAudio generates summary directly from audio file using whisper + audio analysis provider
func (s *Service) GenerateSummaryFromAudio(
	ctx context.Context,
	videoID uuid.UUID,
	audioPath string,
	summaryType string,
	language string,
) (*models.Summary, error) {
	// Get whisper provider to transcribe audio
	whisperProvider, err := s.providerFactory.GetWhisperProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get whisper provider: %w", err)
	}
	
	if whisperProvider == nil {
		return nil, fmt.Errorf("whisper provider not configured for audio analysis")
	}

	// Transcribe audio first
	resp, err := whisperProvider.Transcribe(ctx, whisper.TranscribeRequest{
		AudioPath: audioPath,
		Task:      "transcribe",
	})
	
	// If local whisper fails, try to fallback to Groq
	if err != nil {
		s.logger.Warn("Primary whisper provider failed, attempting Groq fallback", 
			zap.Error(err),
			zap.String("provider", whisperProvider.GetModelInfo().Provider))
		
		// Try to create Groq whisper provider directly as fallback
		if s.config.Whisper.GroqKey != "" {
			groqCfg := whisper.Config{
				Provider: "groq",
				GroqKey:  s.config.Whisper.GroqKey,
			}
			groqProvider, groqErr := whisper.NewProvider(groqCfg)
			if groqErr == nil && groqProvider != nil {
				// Try transcription with Groq
				resp, err = groqProvider.Transcribe(ctx, whisper.TranscribeRequest{
					AudioPath: audioPath,
					Task:      "transcribe",
				})
				if err == nil {
					s.logger.Info("Successfully used Groq whisper as fallback")
				} else {
					return nil, fmt.Errorf("failed to transcribe audio with both primary and Groq fallback: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to create Groq fallback provider: %w", groqErr)
			}
		} else {
			return nil, fmt.Errorf("failed to transcribe audio and Groq key not configured for fallback: %w", err)
		}
	}

	// Get audio analysis provider from settings (for summary generation)
	llmProvider, err := s.providerFactory.GetLLMProvider(ctx, "audio_analysis")
	if err != nil {
		// Fallback to summary provider
		llmProvider, err = s.providerFactory.GetLLMProvider(ctx, "summary")
		if err != nil {
			return nil, fmt.Errorf("failed to get LLM provider: %w", err)
		}
	}

	// Generate summary from transcript using audio analysis provider
	return s.generateSummaryWithProvider(ctx, videoID, resp.Text, summaryType, llmProvider, language)
}

func (s *Service) GenerateSummary(
	ctx context.Context,
	videoID uuid.UUID,
	transcript string,
	summaryType string,
	language string,
) (*models.Summary, error) {
	// Get LLM provider from settings
	llmProvider, err := s.providerFactory.GetLLMProvider(ctx, "summary")
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider: %w", err)
	}
	
	return s.generateSummaryWithProvider(ctx, videoID, transcript, summaryType, llmProvider, language)
}

// generateSummaryWithProvider is a helper method that generates summary with a specific provider
func (s *Service) generateSummaryWithProvider(
	ctx context.Context,
	videoID uuid.UUID,
	transcript string,
	summaryType string,
	llmProvider llm.LLMProvider,
	language string,
) (*models.Summary, error) {
	// Determine summary language: use provided language, or fallback to settings, or default to "auto"
	var summaryLanguage string = "auto"
	if language != "" {
		summaryLanguage = language
	} else if s.settingsService != nil {
		settings, err := s.settingsService.GetSettings(ctx)
		if err == nil && settings != nil && settings.SummaryLanguage != "" {
			summaryLanguage = settings.SummaryLanguage
		}
	}

	// Get prompt template with language
	promptTemplate := prompts.GetSummaryPrompt(summaryType, summaryLanguage)
	prompt := strings.ReplaceAll(promptTemplate, "{{.Transcript}}", transcript)

	// Generate summary using LLM
	req := llm.CompletionRequest{
		Prompt:      prompt,
		SystemPrompt: "You are an expert at analyzing and summarizing video content.",
		MaxTokens:   2000,
		Temperature: 0.7,
		TopP:        0.9,
	}

	resp, err := llmProvider.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	// Parse key points from response
	keyPoints := s.extractKeyPoints(resp.Content)

	// Create summary model
	summary := &models.Summary{
		VideoID:     videoID,
		ModelUsed:   llmProvider.GetModelInfo().Name,
		SummaryType: summaryType,
		Content:     resp.Content,
		KeyPoints:   keyPoints,
	}

	// Save to database
	if err := s.summaryRepo.Create(ctx, summary); err != nil {
		return nil, fmt.Errorf("failed to save summary: %w", err)
	}

	// Record token usage and cost
	if s.costService != nil {
		modelInfo := llmProvider.GetModelInfo()
		inputTokens := resp.InputTokens
		if inputTokens == 0 {
			// Fallback: estimate tokens (rough approximation: 1 token ≈ 4 characters)
			inputTokens = len(prompt) / 4
		}
		outputTokens := resp.OutputTokens
		if outputTokens == 0 && resp.TokensUsed > 0 {
			// Fallback: use total tokens if output tokens not available
			outputTokens = resp.TokensUsed - inputTokens
		}
		
		_ = s.costService.RecordUsage(
			ctx,
			videoID,
			"summarization",
			modelInfo.Provider,
			modelInfo.Name,
			inputTokens,
			outputTokens,
		)
	}

	s.logger.Info("Summary generated", zap.String("video_id", videoID.String()), zap.String("type", summaryType))

	return summary, nil
}

func (s *Service) extractKeyPoints(content string) []string {
	lines := strings.Split(content, "\n")
	var keyPoints []string
	inKeyPointsSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check if we're entering the KEY POINTS section
		if strings.HasPrefix(strings.ToUpper(line), "KEY POINTS") || 
		   strings.HasPrefix(strings.ToUpper(line), "KEY TAKEAWAYS") ||
		   strings.HasPrefix(strings.ToUpper(line), "KEY POINT") {
			inKeyPointsSection = true
			continue
		}
		
		// If we hit another section header (all caps), stop collecting
		if inKeyPointsSection && len(line) > 0 && 
		   strings.ToUpper(line) == line && 
		   !strings.HasPrefix(line, "-") && 
		   !strings.HasPrefix(line, "•") && 
		   !strings.HasPrefix(line, "*") &&
		   !strings.HasPrefix(line, "1.") &&
		   !strings.HasPrefix(line, "2.") &&
		   !strings.HasPrefix(line, "3.") {
			// This might be another section, but check if it's just a short uppercase word
			if len(strings.Fields(line)) > 2 {
				break
			}
		}
		
		// Collect bullet points (various formats)
		if inKeyPointsSection || strings.HasPrefix(line, "-") || 
		   strings.HasPrefix(line, "•") || 
		   strings.HasPrefix(line, "*") ||
		   (len(line) > 2 && (strings.HasPrefix(line, "1.") || 
		                       strings.HasPrefix(line, "2.") || 
		                       strings.HasPrefix(line, "3.") ||
		                       strings.HasPrefix(line, "4.") ||
		                       strings.HasPrefix(line, "5."))) {
			
			// Extract the point text
			point := line
			point = strings.TrimPrefix(point, "-")
			point = strings.TrimPrefix(point, "•")
			point = strings.TrimPrefix(point, "*")
			// Remove numbered prefixes (1., 2., etc.)
			if len(point) > 2 && (point[0] >= '0' && point[0] <= '9' && point[1] == '.') {
				point = strings.TrimSpace(point[2:])
			}
			point = strings.TrimSpace(point)
			
			// Skip empty lines and section headers
			if point != "" && 
			   !strings.HasPrefix(strings.ToUpper(point), "KEY POINTS") &&
			   !strings.HasPrefix(strings.ToUpper(point), "SUMMARY") &&
			   !strings.HasPrefix(strings.ToUpper(point), "TOPICS") {
				keyPoints = append(keyPoints, point)
			}
		}
	}

	// If we didn't find key points in a section, try to extract any bullet points from the entire content
	if len(keyPoints) == 0 {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•") || strings.HasPrefix(line, "*") {
				point := strings.TrimPrefix(line, "-")
				point = strings.TrimPrefix(point, "•")
				point = strings.TrimPrefix(point, "*")
				point = strings.TrimSpace(point)
				if point != "" {
					keyPoints = append(keyPoints, point)
				}
			}
		}
	}

	return keyPoints
}
