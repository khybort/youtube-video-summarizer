package llm

import (
	"context"
)

type LLMProvider interface {
	GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	GenerateCompletionStream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	GetModelInfo() ModelInfo
	ListAvailableModels() ([]string, error)
}

type CompletionRequest struct {
	Prompt      string
	SystemPrompt string
	MaxTokens   int
	Temperature float64
	TopP        float64
}

type CompletionResponse struct {
	Content      string
	TokensUsed   int
	InputTokens  int
	OutputTokens int
	FinishReason string
}

type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

type ModelInfo struct {
	Name        string
	Provider    string
	MaxTokens   int
	SupportsEmbedding bool
}

type Config struct {
	Provider    string
	GeminiKey   string
	GeminiModel string // Empty means auto-detect from API
	OllamaURL   string
	OllamaModel string
}

func NewProvider(cfg Config) (LLMProvider, error) {
	switch cfg.Provider {
	case "gemini":
		return NewGeminiProvider(cfg.GeminiKey, cfg.GeminiModel)
	case "ollama":
		return NewOllamaProvider(cfg.OllamaURL, cfg.OllamaModel)
	default:
		return nil, ErrUnknownProvider
	}
}

