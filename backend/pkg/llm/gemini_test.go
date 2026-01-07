package llm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeminiProvider(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	provider, err := NewGeminiProvider(apiKey, "")
	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestGeminiProvider_GetModelInfo(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	provider, err := NewGeminiProvider(apiKey, "")
	require.NoError(t, err)

	info := provider.GetModelInfo()
	assert.Equal(t, "gemini", info.Provider)
	assert.Equal(t, "gemini-1.5-flash", info.Name)
	assert.True(t, info.SupportsEmbedding)
}

func TestGeminiProvider_GenerateCompletion(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	provider, err := NewGeminiProvider(apiKey, "")
	require.NoError(t, err)

	ctx := context.Background()
	req := CompletionRequest{
		Prompt:      "Say hello in one word",
		SystemPrompt: "",
		MaxTokens:   10,
		Temperature: 0.7,
		TopP:        0.9,
	}

	resp, err := provider.GenerateCompletion(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.Greater(t, resp.TokensUsed, 0)
	assert.Greater(t, resp.InputTokens, 0)
	assert.GreaterOrEqual(t, resp.OutputTokens, 0)
}

func TestGeminiProvider_GenerateEmbedding(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	provider, err := NewGeminiProvider(apiKey, "")
	require.NoError(t, err)

	ctx := context.Background()
	embedding, err := provider.GenerateEmbedding(ctx, "test text")
	require.NoError(t, err)
	assert.NotEmpty(t, embedding)
	assert.Greater(t, len(embedding), 0)
}

