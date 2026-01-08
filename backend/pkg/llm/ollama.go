package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OllamaProvider struct {
	baseURL    string
	httpClient *http.Client
	model      string
}

type ollamaGenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	System  string                 `json:"system,omitempty"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options"`
}

type ollamaGenerateResponse struct {
	Response      string `json:"response"`
	Done          bool   `json:"done"`
	PromptEvalCount int  `json:"prompt_eval_count,omitempty"`
	EvalCount     int    `json:"eval_count,omitempty"`
	TotalDuration int64  `json:"total_duration,omitempty"`
}

type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func NewOllamaProvider(baseURL, model string) (*OllamaProvider, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}

	return &OllamaProvider{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		model:      model,
	}, nil
}

func (o *OllamaProvider) GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	ollamaReq := ollamaGenerateRequest{
		Model:  o.model,
		Prompt: req.Prompt,
		System: req.SystemPrompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": req.Temperature,
			"top_p":       req.TopP,
			"num_predict": req.MaxTokens,
			"num_threads": 2, // Limit CPU threads to prevent excessive CPU usage
		},
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", string(body))
	}

	var ollamaResp ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	inputTokens := ollamaResp.PromptEvalCount
	outputTokens := ollamaResp.EvalCount
	totalTokens := inputTokens + outputTokens

	return &CompletionResponse{
		Content:      ollamaResp.Response,
		TokensUsed:   totalTokens,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		FinishReason: "stop",
	}, nil
}

func (o *OllamaProvider) GenerateCompletionStream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	go func() {
		defer close(ch)

		ollamaReq := ollamaGenerateRequest{
			Model:  o.model,
			Prompt: req.Prompt,
			System: req.SystemPrompt,
			Stream: true,
			Options: map[string]interface{}{
				"temperature": req.Temperature,
				"top_p":       req.TopP,
				"num_predict": req.MaxTokens,
				"num_threads": 2, // Limit CPU threads to prevent excessive CPU usage
			},
		}

		jsonData, _ := json.Marshal(ollamaReq)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
		if err != nil {
			ch <- StreamChunk{Error: err, Done: true}
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := o.httpClient.Do(httpReq)
		if err != nil {
			ch <- StreamChunk{Error: err, Done: true}
			return
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk ollamaGenerateResponse
			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					ch <- StreamChunk{Done: true}
					return
				}
				ch <- StreamChunk{Error: err, Done: true}
				return
			}

			ch <- StreamChunk{Content: chunk.Response, Done: chunk.Done}
			if chunk.Done {
				return
			}
		}
	}()

	return ch, nil
}

func (o *OllamaProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	ollamaReq := ollamaEmbeddingRequest{
		Model:  "nomic-embed-text",
		Prompt: text,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", string(body))
	}

	var ollamaResp ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	return ollamaResp.Embedding, nil
}

func (o *OllamaProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := o.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (o *OllamaProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:            o.model,
		Provider:        "ollama",
		MaxTokens:       4096,
		SupportsEmbedding: true,
	}
}

func (o *OllamaProvider) ListAvailableModels() ([]string, error) {
	req, err := http.NewRequest("GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		// Remove tag prefix if present
		models[i] = strings.TrimPrefix(m.Name, o.model+":")
	}

	return models, nil
}

