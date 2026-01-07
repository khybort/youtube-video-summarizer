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

type LocalWhisperProvider struct {
	baseURL    string
	httpClient *http.Client
}

func NewLocalWhisperProvider(baseURL string) (*LocalWhisperProvider, error) {
	if baseURL == "" {
		baseURL = "http://localhost:8001"
	}

	return &LocalWhisperProvider{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // Whisper transcription can take time for long videos
		},
	}, nil
}

func (l *LocalWhisperProvider) Transcribe(ctx context.Context, req TranscribeRequest) (*TranscribeResponse, error) {
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

	if req.Language != "" {
		writer.WriteField("language", req.Language)
	}
	writer.WriteField("task", req.Task)

	writer.Close()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", l.baseURL+"/transcribe", body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := l.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("local whisper API error: %s", string(bodyBytes))
	}

	var result struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Duration float64 `json:"duration"`
		Segments []struct {
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
			ID:    i,
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

func (l *LocalWhisperProvider) GetSupportedLanguages() []string {
	return []string{"auto"} // Local whisper supports all languages
}

func (l *LocalWhisperProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:     "faster-whisper",
		Provider: "local",
	}
}

