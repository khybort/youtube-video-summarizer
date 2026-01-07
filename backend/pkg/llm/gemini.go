package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client    *genai.Client
	model     *genai.GenerativeModel
	modelName string
}

// findAvailableModel tries to find an available model from the API
func findAvailableModel(ctx context.Context, client *genai.Client) (string, error) {
	// Priority order: try newer models first
	preferredModels := []string{
		"gemini-1.5-flash",
		"gemini-1.5-pro",
		"gemini-pro",
	}

	// Get list of available models from API
	iter := client.ListModels(ctx)
	availableModels := make(map[string]bool)
	availableModelNames := []string{}
	
	for {
		model, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// If we can't list models, fallback to trying preferred models
			break
		}
		
		// Extract model name (format: "models/gemini-1.5-flash" -> "gemini-1.5-flash")
		modelName := strings.TrimPrefix(model.Name, "models/")
		availableModels[modelName] = true
		availableModelNames = append(availableModelNames, modelName)
		
		// Check if model supports generateContent
		for _, method := range model.SupportedGenerationMethods {
			if method == "generateContent" {
				// This model supports generation, use it if it's in our preferred list
				for _, preferred := range preferredModels {
					if modelName == preferred {
						return modelName, nil
					}
				}
			}
		}
	}

	// If we found available models, try preferred ones in order
	if len(availableModels) > 0 {
		for _, modelName := range preferredModels {
			if availableModels[modelName] {
				return modelName, nil
			}
		}
		// If none of preferred models are available, use first available one
		if len(availableModelNames) > 0 {
			return availableModelNames[0], nil
		}
	}

	// If we couldn't list models, try preferred models anyway
	// The API will return an error if the model doesn't exist
	return preferredModels[0], nil
}

func NewGeminiProvider(apiKey string, modelNameOverride string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	var modelName string
	if modelNameOverride != "" {
		// Use user-specified model
		modelName = modelNameOverride
	} else {
		// Dynamically find available model from API
		modelName, err = findAvailableModel(ctx, client)
		if err != nil {
			// Fallback to default if we can't determine available models
			modelName = "gemini-1.5-flash"
		}
	}

	model := client.GenerativeModel(modelName)

	return &GeminiProvider{
		client:    client,
		model:     model,
		modelName: modelName,
	}, nil
}

func (g *GeminiProvider) GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	prompt := req.Prompt
	if req.SystemPrompt != "" {
		prompt = req.SystemPrompt + "\n\n" + prompt
	}

	// Configure generation config
	temp := float32(req.Temperature)
	topP := float32(req.TopP)
	maxTokens := int32(req.MaxTokens)
	
	g.model.Temperature = &temp
	g.model.TopP = &topP
	g.model.MaxOutputTokens = &maxTokens

	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	candidate := resp.Candidates[0]
	content := ""
	for _, part := range candidate.Content.Parts {
		if text, ok := part.(genai.Text); ok {
			content += string(text)
		}
	}

	// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
	// Gemini API doesn't always return UsageMetadata in all versions
	tokensUsed := len(prompt) / 4 + len(content) / 4
	inputTokens := len(prompt) / 4
	outputTokens := len(content) / 4

	return &CompletionResponse{
		Content:      content,
		TokensUsed:   tokensUsed,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		FinishReason: candidate.FinishReason.String(),
	}, nil
}

func (g *GeminiProvider) GenerateCompletionStream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	go func() {
		defer close(ch)

		prompt := req.Prompt
		if req.SystemPrompt != "" {
			prompt = req.SystemPrompt + "\n\n" + prompt
		}

		iter := g.model.GenerateContentStream(ctx, genai.Text(prompt))
		for {
			resp, err := iter.Next()
			if err != nil {
				ch <- StreamChunk{Error: err, Done: true}
				return
			}

			if len(resp.Candidates) > 0 {
				candidate := resp.Candidates[0]
				for _, part := range candidate.Content.Parts {
					if text, ok := part.(genai.Text); ok {
						ch <- StreamChunk{Content: string(text), Done: false}
					}
				}
			}
		}
	}()

	return ch, nil
}

func (g *GeminiProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	em := g.client.EmbeddingModel("text-embedding-004")
	res, err := em.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("embedding generation failed: %w", err)
	}

	return res.Embedding.Values, nil
}

func (g *GeminiProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := g.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (g *GeminiProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:             g.modelName,
		Provider:         "gemini",
		MaxTokens:         8192,
		SupportsEmbedding: true,
	}
}

func (g *GeminiProvider) ListAvailableModels() ([]string, error) {
	ctx := context.Background()
	iter := g.client.ListModels(ctx)
	
	var models []string
	for {
		model, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// If we can't list models, return common fallback list
			return []string{"gemini-1.5-flash", "gemini-1.5-pro", "gemini-pro"}, err
		}
		
		// Extract model name (format: "models/gemini-1.5-flash" -> "gemini-1.5-flash")
		modelName := strings.TrimPrefix(model.Name, "models/")
		models = append(models, modelName)
	}
	
	if len(models) == 0 {
		// Fallback if no models found
		return []string{"gemini-1.5-flash", "gemini-1.5-pro", "gemini-pro"}, nil
	}
	
	return models, nil
}

