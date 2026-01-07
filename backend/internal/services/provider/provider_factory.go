package provider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/config"
	settingsservice "youtube-video-summarizer/backend/internal/services/settings"
	"youtube-video-summarizer/backend/pkg/llm"
	"youtube-video-summarizer/backend/pkg/whisper"
)

// ProviderFactory manages LLM and Whisper providers based on settings
type ProviderFactory struct {
	settingsService *settingsservice.Service
	config          *config.Config
	logger          *zap.Logger
	
	// Cache for providers
	llmCache      map[string]llm.LLMProvider
	whisperCache  map[string]whisper.WhisperProvider
	cacheMutex    sync.RWMutex
}

func NewProviderFactory(
	settingsService *settingsservice.Service,
	cfg *config.Config,
	logger *zap.Logger,
) *ProviderFactory {
	return &ProviderFactory{
		settingsService: settingsService,
		config:          cfg,
		logger:          logger,
		llmCache:       make(map[string]llm.LLMProvider),
		whisperCache:   make(map[string]whisper.WhisperProvider),
	}
}

// GetLLMProvider returns an LLM provider based on settings for the given operation
func (f *ProviderFactory) GetLLMProvider(ctx context.Context, operation string) (llm.LLMProvider, error) {
	settings, err := f.settingsService.GetSettings(ctx)
	if err != nil {
		f.logger.Warn("Failed to get settings, using default", zap.Error(err))
		// Fallback to default from config
		return f.getDefaultLLMProvider()
	}

	var providerName string
	var ollamaURL string
	var ollamaModel string

	switch operation {
	case "summary":
		providerName = settings.SummaryProvider
	case "embedding":
		providerName = settings.EmbeddingProvider
	case "audio_analysis":
		providerName = settings.AudioAnalysisProvider
	default:
		providerName = settings.SummaryProvider // Default to summary provider
	}

	// Get Ollama settings from settings or config
	if settings.OllamaURL != "" {
		ollamaURL = settings.OllamaURL
	} else {
		ollamaURL = f.config.LLM.OllamaURL
	}

	if settings.OllamaModel != "" {
		ollamaModel = settings.OllamaModel
	} else {
		ollamaModel = f.config.LLM.OllamaModel
	}

	// Get Gemini model from settings (if specified)
	var geminiModel string
	if settings.GeminiModel != "" {
		geminiModel = settings.GeminiModel
	}

	// Prefer DB-stored Gemini key; fall back to env/config.
	geminiKey := f.config.LLM.GeminiKey
	if settings.GeminiAPIKey != "" {
		geminiKey = settings.GeminiAPIKey
	}

	// Create cache key (include gemini model + key hash so changes take effect without restart)
	geminiHash := fmt.Sprintf("%x", sha256.Sum256([]byte(geminiKey)))[:8]
	cacheKey := fmt.Sprintf("%s:%s:%s:%s:%s", providerName, ollamaURL, ollamaModel, geminiModel, geminiHash)

	// Check cache
	f.cacheMutex.RLock()
	if provider, ok := f.llmCache[cacheKey]; ok {
		f.cacheMutex.RUnlock()
		return provider, nil
	}
	f.cacheMutex.RUnlock()

	// Create new provider
	cfg := llm.Config{
		Provider:    providerName,
		GeminiKey:   geminiKey,
		GeminiModel: geminiModel,
		OllamaURL:   ollamaURL,
		OllamaModel: ollamaModel,
	}

	provider, err := llm.NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// Cache it
	f.cacheMutex.Lock()
	f.llmCache[cacheKey] = provider
	f.cacheMutex.Unlock()

	return provider, nil
}

// GetWhisperProvider returns a Whisper provider based on settings
func (f *ProviderFactory) GetWhisperProvider(ctx context.Context) (whisper.WhisperProvider, error) {
	settings, err := f.settingsService.GetSettings(ctx)
	if err != nil {
		f.logger.Warn("Failed to get settings, using default", zap.Error(err))
		// Fallback to default from config
		return f.getDefaultWhisperProvider()
	}

	providerName := settings.TranscriptProvider

	// YouTube doesn't need a Whisper provider, return nil
	if providerName == "youtube" {
		return nil, nil
	}

	// Get settings from settings or config
	var groqKey string
	var huggingFaceKey string
	var localWhisperURL string
	var whisperModel string

	// Prefer DB-stored keys; fall back to env/config keys.
	if settings.GroqAPIKey != "" {
		groqKey = settings.GroqAPIKey
	} else if f.config.Whisper.GroqKey != "" {
		groqKey = f.config.Whisper.GroqKey
	}

	if settings.HuggingFaceAPIKey != "" {
		huggingFaceKey = settings.HuggingFaceAPIKey
	} else if f.config.Whisper.HuggingFaceKey != "" {
		huggingFaceKey = f.config.Whisper.HuggingFaceKey
	}

	if settings.LocalWhisperURL != "" {
		localWhisperURL = settings.LocalWhisperURL
	} else {
		localWhisperURL = f.config.Whisper.LocalWhisperURL
	}

	if settings.WhisperModel != "" {
		whisperModel = settings.WhisperModel
	} else {
		whisperModel = f.config.Whisper.LocalModel
	}

	// Create cache key (include a short hash of API keys so key changes take effect without restarting)
	groqHash := fmt.Sprintf("%x", sha256.Sum256([]byte(groqKey)))[:8]
	hfHash := fmt.Sprintf("%x", sha256.Sum256([]byte(huggingFaceKey)))[:8]
	cacheKey := fmt.Sprintf("%s:%s:%s:%s:%s", providerName, localWhisperURL, whisperModel, groqHash, hfHash)

	// Check cache
	f.cacheMutex.RLock()
	if provider, ok := f.whisperCache[cacheKey]; ok {
		f.cacheMutex.RUnlock()
		return provider, nil
	}
	f.cacheMutex.RUnlock()

	// Create new provider
	cfg := whisper.Config{
		Provider:        providerName,
		GroqKey:         groqKey,
		HuggingFaceKey:  huggingFaceKey,
		LocalWhisperURL: localWhisperURL,
		LocalModel:      whisperModel,
	}

	provider, err := whisper.NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Whisper provider: %w", err)
	}

	// Cache it
	f.cacheMutex.Lock()
	f.whisperCache[cacheKey] = provider
	f.cacheMutex.Unlock()

	return provider, nil
}

// getDefaultLLMProvider returns the default LLM provider from config
func (f *ProviderFactory) getDefaultLLMProvider() (llm.LLMProvider, error) {
	cfg := llm.Config{
		Provider:    f.config.LLM.Provider,
		GeminiKey:   f.config.LLM.GeminiKey,
		GeminiModel: "", // Auto-detect
		OllamaURL:   f.config.LLM.OllamaURL,
		OllamaModel: f.config.LLM.OllamaModel,
	}
	return llm.NewProvider(cfg)
}

// getDefaultWhisperProvider returns the default Whisper provider from config
func (f *ProviderFactory) getDefaultWhisperProvider() (whisper.WhisperProvider, error) {
	cfg := whisper.Config{
		Provider:        f.config.Whisper.Provider,
		GroqKey:         f.config.Whisper.GroqKey,
		HuggingFaceKey:  f.config.Whisper.HuggingFaceKey,
		LocalWhisperURL: f.config.Whisper.LocalWhisperURL,
		LocalModel:      f.config.Whisper.LocalModel,
	}
	return whisper.NewProvider(cfg)
}

// ClearCache clears the provider cache (useful when settings change)
func (f *ProviderFactory) ClearCache() {
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()
	f.llmCache = make(map[string]llm.LLMProvider)
	f.whisperCache = make(map[string]whisper.WhisperProvider)
}

