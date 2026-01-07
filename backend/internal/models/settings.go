package models

import (
	"time"

	"github.com/google/uuid"
)

// Settings stores user preferences for model selection per operation
type Settings struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	
	// Model selections for different operations
	TranscriptProvider   string `gorm:"type:varchar(50);default:'youtube'" json:"transcript_provider"` // youtube, groq, local, huggingface
	SummaryProvider      string `gorm:"type:varchar(50);default:'gemini'" json:"summary_provider"`     // gemini, ollama
	EmbeddingProvider    string `gorm:"type:varchar(50);default:'gemini'" json:"embedding_provider"`   // gemini, ollama
	AudioAnalysisProvider string `gorm:"type:varchar(50);default:'gemini'" json:"audio_analysis_provider"` // gemini, ollama (for future audio analysis)
	
	// Model-specific settings
	OllamaModel          string `gorm:"type:varchar(100);default:'llama3.2'" json:"ollama_model"`
	WhisperModel         string `gorm:"type:varchar(100);default:'base'" json:"whisper_model"`
	GeminiModel          string `gorm:"type:varchar(100)" json:"gemini_model"` // Empty means auto-detect
	
	// Provider URLs (can override defaults)
	OllamaURL            string `gorm:"type:varchar(255)" json:"ollama_url"`
	LocalWhisperURL      string `gorm:"type:varchar(255)" json:"local_whisper_url"`

	// Summary language preference
	SummaryLanguage      string `gorm:"type:varchar(50);default:'auto'" json:"summary_language"` // auto, en, tr, es, fr, de, etc.

	// Provider API keys (stored server-side; not returned in GET /settings)
	GroqAPIKey          string `gorm:"type:varchar(255)" json:"-"`
	HuggingFaceAPIKey   string `gorm:"type:varchar(255)" json:"-"`
	GeminiAPIKey        string `gorm:"type:varchar(255)" json:"-"`
	
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Settings) TableName() string {
	return "settings"
}

// GetDefaultSettings returns default settings
func GetDefaultSettings() *Settings {
	return &Settings{
		TranscriptProvider:   "youtube",
		SummaryProvider:      "gemini",
		EmbeddingProvider:    "gemini",
		AudioAnalysisProvider: "gemini",
		OllamaModel:          "llama3.2",
		WhisperModel:         "base",
		GeminiModel:          "", // Empty means auto-detect from API
		OllamaURL:            "http://localhost:11434",
		LocalWhisperURL:      "http://localhost:8001",
		SummaryLanguage:      "auto", // auto means use transcript language
	}
}

