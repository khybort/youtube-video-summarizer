package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type HuggingFaceWhisperProvider struct {
	apiKey     string
	httpClient *http.Client
	model      string
	baseURL    string
}

func NewHuggingFaceWhisperProvider(apiKey string) (*HuggingFaceWhisperProvider, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}

	return &HuggingFaceWhisperProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Hugging Face can take longer for large files
		},
		model:   "openai/whisper-large-v3",
		baseURL: "https://api-inference.huggingface.co",
	}, nil
}

func (h *HuggingFaceWhisperProvider) Transcribe(ctx context.Context, req TranscribeRequest) (*TranscribeResponse, error) {
	var audioData []byte
	var err error

	if req.AudioPath != "" {
		audioData, err = os.ReadFile(req.AudioPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read audio file: %w", err)
		}
	} else if len(req.AudioData) > 0 {
		audioData = req.AudioData
	} else {
		return nil, fmt.Errorf("no audio data provided")
	}

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("inputs", "audio.mp3")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(audioData); err != nil {
		return nil, err
	}

	// Add parameters if needed
	params := map[string]interface{}{
		"return_timestamps": true,
	}

	if req.Language != "" {
		params["language"] = req.Language
	}

	// Add parameters as JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	writer.WriteField("parameters", string(paramsJSON))
	writer.Close()

	// Make request to Hugging Face Inference API
	url := fmt.Sprintf("%s/models/%s", h.baseURL, h.model)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+h.apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// Handle different response statuses
	if resp.StatusCode == http.StatusServiceUnavailable {
		// Model is loading, return error with retry info
		var loadingResp struct {
			Error string `json:"error"`
			EstimatedTime float64 `json:"estimated_time,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &loadingResp); err == nil {
			return nil, fmt.Errorf("model is loading, estimated time: %.0f seconds. Please try again later", loadingResp.EstimatedTime)
		}
		return nil, fmt.Errorf("model is loading, please try again later")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hugging face API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response - Hugging Face can return different formats
	var result struct {
		Text     string `json:"text"`
		Chunks   []struct {
			Text      string    `json:"text"`
			Timestamp []float64 `json:"timestamp"`
		} `json:"chunks,omitempty"`
		// Alternative format with raw text
		RawText string `json:"raw_text,omitempty"`
	}

	// Try to parse as JSON first
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		// If JSON parsing fails, try as plain text
		text := string(bodyBytes)
		// Remove quotes if present
		if len(text) > 2 && text[0] == '"' && text[len(text)-1] == '"' {
			text = text[1 : len(text)-1]
		}
		if len(text) > 0 {
			result.Text = text
		} else {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
	}

	// Use RawText if Text is empty
	if result.Text == "" && result.RawText != "" {
		result.Text = result.RawText
	}

	// Convert chunks to segments
	segments := make([]TranscriptSegment, 0)
	if len(result.Chunks) > 0 {
		for i, chunk := range result.Chunks {
			start := 0.0
			end := 0.0
			if len(chunk.Timestamp) >= 2 {
				start = chunk.Timestamp[0]
				end = chunk.Timestamp[1]
			}
			segments = append(segments, TranscriptSegment{
				ID:    i,
				Start: start,
				End:   end,
				Text:  chunk.Text,
			})
		}
	} else if result.Text != "" {
		// If no chunks, create a single segment with the full text
		segments = append(segments, TranscriptSegment{
			ID:    0,
			Start: 0.0,
			End:   0.0, // Duration unknown
			Text:  result.Text,
		})
	}

	// Try to extract language from response if available
	language := req.Language
	if language == "" {
		language = "auto"
	}

	return &TranscribeResponse{
		Text:     result.Text,
		Language: language,
		Duration: 0.0, // Hugging Face doesn't always return duration
		Segments: segments,
	}, nil
}

func (h *HuggingFaceWhisperProvider) GetSupportedLanguages() []string {
	return []string{"auto", "en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh", "tr", "ar", "hi", "nl", "pl", "sv", "no", "da", "fi", "el", "cs", "ro", "hu", "bg", "hr", "sk", "sl", "et", "lv", "lt", "mt", "ga", "cy"}
}

func (h *HuggingFaceWhisperProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:     h.model,
		Provider: "huggingface",
	}
}

