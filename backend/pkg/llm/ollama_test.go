package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOllamaProvider(t *testing.T) {
	provider, err := NewOllamaProvider("", "")
	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestOllamaProvider_GetModelInfo(t *testing.T) {
	provider, err := NewOllamaProvider("", "llama3.2")
	require.NoError(t, err)

	info := provider.GetModelInfo()
	assert.Equal(t, "ollama", info.Provider)
	assert.Equal(t, "llama3.2", info.Name)
}

func TestOllamaProvider_GenerateCompletion(t *testing.T) {
	// This test requires Ollama to be running
	provider, err := NewOllamaProvider("", "llama3.2")
	require.NoError(t, err)

	ctx := context.Background()
	req := CompletionRequest{
		Prompt:      "Say hello",
		SystemPrompt: "",
		MaxTokens:   10,
		Temperature: 0.7,
		TopP:        0.9,
	}

	// Skip if Ollama is not available
	resp, err := provider.GenerateCompletion(ctx, req)
	if err != nil {
		t.Skip("Ollama not available:", err)
	}

	assert.NotEmpty(t, resp.Content)
	// Ollama may return 0 tokens if not available in response
	assert.GreaterOrEqual(t, resp.TokensUsed, 0)
}

