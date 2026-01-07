package llm

import "errors"

var (
	ErrUnknownProvider     = errors.New("unknown LLM provider")
	ErrProviderUnavailable = errors.New("provider unavailable")
	ErrRateLimited         = errors.New("rate limited")
	ErrInvalidAPIKey       = errors.New("invalid API key")
	ErrModelNotFound       = errors.New("model not found")
)

