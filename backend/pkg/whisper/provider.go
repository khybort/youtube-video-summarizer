package whisper

import (
	"context"
)

type WhisperProvider interface {
	Transcribe(ctx context.Context, req TranscribeRequest) (*TranscribeResponse, error)
	GetSupportedLanguages() []string
	GetModelInfo() ModelInfo
}

type TranscribeRequest struct {
	AudioPath   string
	AudioData   []byte
	Language    string
	Task        string // "transcribe" or "translate"
	WordTimings bool
}

type TranscribeResponse struct {
	Text     string
	Language string
	Duration float64
	Segments []TranscriptSegment
}

type TranscriptSegment struct {
	ID    int
	Start float64
	End   float64
	Text  string
}

type ModelInfo struct {
	Name string
	Provider string
}

type Config struct {
	Provider        string
	GroqKey         string
	HuggingFaceKey  string
	LocalWhisperURL string
	LocalModel      string
}

func NewProvider(cfg Config) (WhisperProvider, error) {
	switch cfg.Provider {
	case "groq":
		return NewGroqWhisperProvider(cfg.GroqKey)
	case "huggingface":
		return NewHuggingFaceWhisperProvider(cfg.HuggingFaceKey)
	case "local":
		return NewLocalWhisperProvider(cfg.LocalWhisperURL)
	default:
		return nil, ErrUnknownProvider
	}
}

