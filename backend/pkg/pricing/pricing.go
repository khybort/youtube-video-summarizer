package pricing

import (
	"fmt"
)

// Pricing data for different providers and models
// Prices are per 1K tokens (as of 2024)
var PricingData = map[string]map[string]Pricing{
	"gemini": {
		"gemini-1.5-flash": {
			InputCostPer1K:  0.000075,  // $0.075 per 1M tokens
			OutputCostPer1K: 0.0003,    // $0.30 per 1M tokens
		},
		"gemini-1.5-pro": {
			InputCostPer1K:  0.00125,   // $1.25 per 1M tokens
			OutputCostPer1K: 0.005,     // $5.00 per 1M tokens
		},
	},
	"ollama": {
		"llama3.2": {
			InputCostPer1K:  0.0, // Local, no cost
			OutputCostPer1K: 0.0,
		},
		"llama3.1": {
			InputCostPer1K:  0.0,
			OutputCostPer1K: 0.0,
		},
		"mistral": {
			InputCostPer1K:  0.0,
			OutputCostPer1K: 0.0,
		},
	},
	"groq": {
		"whisper-large-v3": {
			InputCostPer1K:  0.0, // Groq Whisper is typically per minute, not per token
			OutputCostPer1K: 0.0,
		},
	},
	"local": {
		"faster-whisper": {
			InputCostPer1K:  0.0, // Local, no cost
			OutputCostPer1K: 0.0,
		},
	},
}

type Pricing struct {
	InputCostPer1K  float64
	OutputCostPer1K float64
}

func GetPricing(provider, model string) (Pricing, error) {
	providerPricing, ok := PricingData[provider]
	if !ok {
		return Pricing{}, fmt.Errorf("unknown provider: %s", provider)
	}

	pricing, ok := providerPricing[model]
	if !ok {
		// Return default pricing if model not found
		return Pricing{
			InputCostPer1K:  0.0,
			OutputCostPer1K: 0.0,
		}, nil
	}

	return pricing, nil
}

func CalculateCost(provider, model string, inputTokens, outputTokens int) (float64, error) {
	pricing, err := GetPricing(provider, model)
	if err != nil {
		return 0, err
	}

	inputCost := (float64(inputTokens) / 1000.0) * pricing.InputCostPer1K
	outputCost := (float64(outputTokens) / 1000.0) * pricing.OutputCostPer1K

	return inputCost + outputCost, nil
}

// Special handling for Groq Whisper (per minute pricing)
func CalculateGroqWhisperCost(durationMinutes float64) float64 {
	// Groq Whisper pricing: $0.006 per minute (approximate)
	// This is a simplified calculation
	return durationMinutes * 0.006
}

