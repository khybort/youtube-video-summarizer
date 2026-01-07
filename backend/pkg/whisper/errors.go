package whisper

import "errors"

var (
	ErrUnknownProvider     = errors.New("unknown Whisper provider")
	ErrProviderUnavailable = errors.New("provider unavailable")
	ErrTranscriptionFailed = errors.New("transcription failed")
	ErrAudioTooLong        = errors.New("audio exceeds maximum duration")
	ErrInvalidAPIKey       = errors.New("invalid API key")
)

