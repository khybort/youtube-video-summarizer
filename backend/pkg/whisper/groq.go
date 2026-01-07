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
)

type GroqWhisperProvider struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

func NewGroqWhisperProvider(apiKey string) (*GroqWhisperProvider, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}

	return &GroqWhisperProvider{
		apiKey: apiKey,
		httpClient: &http.Client{},
		model:      "whisper-large-v3",
	}, nil
}

func (g *GroqWhisperProvider) Transcribe(ctx context.Context, req TranscribeRequest) (*TranscribeResponse, error) {
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

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(audioData); err != nil {
		return nil, err
	}

	// Add model
	writer.WriteField("model", g.model)
	writer.WriteField("response_format", "verbose_json")
	writer.WriteField("timestamp_granularities[]", "segment")

	if req.Language != "" {
		writer.WriteField("language", req.Language)
	}

	writer.Close()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("groq API error: %s", string(bodyBytes))
	}

	var result struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Duration float64 `json:"duration"`
		Segments []struct {
			ID    int     `json:"id"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
		} `json:"segments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	segments := make([]TranscriptSegment, len(result.Segments))
	for i, seg := range result.Segments {
		segments[i] = TranscriptSegment{
			ID:    seg.ID,
			Start: seg.Start,
			End:   seg.End,
			Text:  seg.Text,
		}
	}

	return &TranscribeResponse{
		Text:     result.Text,
		Language: result.Language,
		Duration: result.Duration,
		Segments: segments,
	}, nil
}

func (g *GroqWhisperProvider) GetSupportedLanguages() []string {
	return []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh", "tr"}
}

func (g *GroqWhisperProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:     g.model,
		Provider: "groq",
	}
}

