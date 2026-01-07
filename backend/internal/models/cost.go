package models

import (
	"time"

	"github.com/google/uuid"
)

type TokenUsage struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID     uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	Operation   string    `gorm:"type:varchar(50);not null;index" json:"operation"` // "transcription", "summarization", "embedding"
	Provider    string    `gorm:"type:varchar(50);not null;index" json:"provider"`   // "gemini", "ollama", "groq", "local"
	Model       string    `gorm:"type:varchar(100);not null" json:"model"`
	InputTokens int       `gorm:"default:0" json:"input_tokens"`
	OutputTokens int     `gorm:"default:0" json:"output_tokens"`
	TotalTokens int      `gorm:"default:0" json:"total_tokens"`
	Cost        float64   `gorm:"type:decimal(10,6);default:0" json:"cost"` // USD
	CreatedAt   time.Time `gorm:"autoCreateTime;index:idx_token_usage_created_at" json:"created_at"`
	Video       Video     `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE" json:"-"`
}

func (TokenUsage) TableName() string {
	return "token_usage"
}

type CostSummary struct {
	TotalCost      float64            `json:"total_cost"`
	TotalTokens    int                `json:"total_tokens"`
	ByProvider     map[string]float64 `json:"by_provider"`
	ByOperation    map[string]float64 `json:"by_operation"`
	ByModel        map[string]float64 `json:"by_model"`
	Period         string             `json:"period"` // "today", "week", "month", "all"
	VideoCount     int                `json:"video_count"`
	AverageCostPerVideo float64       `json:"average_cost_per_video"`
}

type PricingConfig struct {
	Provider string             `json:"provider"`
	Models   map[string]Pricing `json:"models"`
}

type Pricing struct {
	InputCostPer1K  float64 `json:"input_cost_per_1k"`  // Cost per 1K input tokens
	OutputCostPer1K float64 `json:"output_cost_per_1k"` // Cost per 1K output tokens
}
