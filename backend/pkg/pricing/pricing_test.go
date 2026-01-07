package pricing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPricing(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		wantErr  bool
	}{
		{
			name:     "valid gemini flash",
			provider: "gemini",
			model:    "gemini-1.5-flash",
			wantErr:  false,
		},
		{
			name:     "valid gemini pro",
			provider: "gemini",
			model:    "gemini-1.5-pro",
			wantErr:  false,
		},
		{
			name:     "valid ollama",
			provider: "ollama",
			model:    "llama3.2",
			wantErr:  false,
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			model:    "model",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing, err := GetPricing(tt.provider, tt.model)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pricing)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		model        string
		inputTokens  int
		outputTokens int
		wantCost     float64
		wantErr      bool
	}{
		{
			name:         "gemini flash calculation",
			provider:     "gemini",
			model:        "gemini-1.5-flash",
			inputTokens:  1000,
			outputTokens: 500,
			wantCost:     0.000075 + 0.00015, // (1000/1000)*0.000075 + (500/1000)*0.0003
			wantErr:      false,
		},
		{
			name:         "ollama free",
			provider:     "ollama",
			model:        "llama3.2",
			inputTokens:  1000,
			outputTokens: 500,
			wantCost:     0.0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := CalculateCost(tt.provider, tt.model, tt.inputTokens, tt.outputTokens)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.wantCost, cost, 0.000001)
			}
		})
	}
}

func TestCalculateGroqWhisperCost(t *testing.T) {
	tests := []struct {
		name             string
		durationMinutes  float64
		expectedCost     float64
	}{
		{
			name:            "1 minute",
			durationMinutes: 1.0,
			expectedCost:    0.006,
		},
		{
			name:            "10 minutes",
			durationMinutes: 10.0,
			expectedCost:    0.06,
		},
		{
			name:            "zero minutes",
			durationMinutes: 0.0,
			expectedCost:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := CalculateGroqWhisperCost(tt.durationMinutes)
			assert.InDelta(t, tt.expectedCost, cost, 0.0001)
		})
	}
}

