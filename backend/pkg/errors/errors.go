package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents the main error category
type ErrorCode string

const (
	// General errors
	ErrorCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrorCodeConflict     ErrorCode = "CONFLICT"
	ErrorCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrorCodeTimeout      ErrorCode = "TIMEOUT"
	ErrorCodeTooLarge     ErrorCode = "PAYLOAD_TOO_LARGE"

	// Video errors
	ErrorCodeVideoNotFound      ErrorCode = "VIDEO_NOT_FOUND"
	ErrorCodeVideoInvalidURL    ErrorCode = "VIDEO_INVALID_URL"
	ErrorCodeVideoAlreadyExists ErrorCode = "VIDEO_ALREADY_EXISTS"
	ErrorCodeVideoProcessing    ErrorCode = "VIDEO_PROCESSING_ERROR"

	// Transcript errors
	ErrorCodeTranscriptNotFound      ErrorCode = "TRANSCRIPT_NOT_FOUND"
	ErrorCodeTranscriptGeneration    ErrorCode = "TRANSCRIPT_GENERATION_ERROR"
	ErrorCodeTranscriptFileTooLarge  ErrorCode = "TRANSCRIPT_FILE_TOO_LARGE"
	ErrorCodeTranscriptNoCaptions    ErrorCode = "TRANSCRIPT_NO_CAPTIONS_AVAILABLE"

	// Summary errors
	ErrorCodeSummaryNotFound     ErrorCode = "SUMMARY_NOT_FOUND"
	ErrorCodeSummaryGeneration   ErrorCode = "SUMMARY_GENERATION_ERROR"
	ErrorCodeSummaryNoTranscript ErrorCode = "SUMMARY_NO_TRANSCRIPT"

	// Embedding errors
	ErrorCodeEmbeddingNotFound   ErrorCode = "EMBEDDING_NOT_FOUND"
	ErrorCodeEmbeddingGeneration ErrorCode = "EMBEDDING_GENERATION_ERROR"

	// Similarity errors
	ErrorCodeSimilarityNotFound      ErrorCode = "SIMILARITY_NOT_FOUND"
	ErrorCodeSimilarityCalculation   ErrorCode = "SIMILARITY_CALCULATION_ERROR"
	ErrorCodeSimilarityNoEmbeddings ErrorCode = "SIMILARITY_NO_EMBEDDINGS"

	// Provider errors
	ErrorCodeProviderNotFound      ErrorCode = "PROVIDER_NOT_FOUND"
	ErrorCodeProviderUnavailable    ErrorCode = "PROVIDER_UNAVAILABLE"
	ErrorCodeProviderConfiguration  ErrorCode = "PROVIDER_CONFIGURATION_ERROR"
	ErrorCodeProviderRateLimit      ErrorCode = "PROVIDER_RATE_LIMIT"
	ErrorCodeProviderQuotaExceeded ErrorCode = "PROVIDER_QUOTA_EXCEEDED"

	// Database errors
	ErrorCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION_ERROR"
	ErrorCodeDatabaseQuery      ErrorCode = "DATABASE_QUERY_ERROR"
	ErrorCodeDatabaseConstraint ErrorCode = "DATABASE_CONSTRAINT_ERROR"

	// External service errors
	ErrorCodeYouTubeAPI      ErrorCode = "YOUTUBE_API_ERROR"
	ErrorCodeYouTubeDownload ErrorCode = "YOUTUBE_DOWNLOAD_ERROR"
	ErrorCodeLLMAPI          ErrorCode = "LLM_API_ERROR"
	ErrorCodeWhisperAPI      ErrorCode = "WHISPER_API_ERROR"
)

// SubCode represents a more specific error within a category
type SubCode string

const (
	// General subcodes
	SubCodeUnknown          SubCode = "UNKNOWN"
	SubCodeInvalidInput     SubCode = "INVALID_INPUT"
	SubCodeMissingParameter SubCode = "MISSING_PARAMETER"
	SubCodeInvalidFormat    SubCode = "INVALID_FORMAT"

	// Video subcodes
	SubCodeVideoIDInvalid     SubCode = "VIDEO_ID_INVALID"
	SubCodeVideoURLInvalid    SubCode = "VIDEO_URL_INVALID"
	SubCodeVideoNotFound      SubCode = "VIDEO_NOT_FOUND"
	SubCodeVideoDuplicate     SubCode = "VIDEO_DUPLICATE"
	SubCodeVideoStatusInvalid SubCode = "VIDEO_STATUS_INVALID"

	// Transcript subcodes
	SubCodeTranscriptNotFound      SubCode = "TRANSCRIPT_NOT_FOUND"
	SubCodeTranscriptDownloadFailed SubCode = "TRANSCRIPT_DOWNLOAD_FAILED"
	SubCodeTranscriptParseFailed    SubCode = "TRANSCRIPT_PARSE_FAILED"
	SubCodeTranscriptFileTooLarge   SubCode = "TRANSCRIPT_FILE_TOO_LARGE"
	SubCodeTranscriptNoCaptions     SubCode = "TRANSCRIPT_NO_CAPTIONS"
	SubCodeTranscriptWhisperFailed  SubCode = "TRANSCRIPT_WHISPER_FAILED"

	// Summary subcodes
	SubCodeSummaryNotFound      SubCode = "SUMMARY_NOT_FOUND"
	SubCodeSummaryLLMFailed     SubCode = "SUMMARY_LLM_FAILED"
	SubCodeSummaryNoTranscript  SubCode = "SUMMARY_NO_TRANSCRIPT"
	SubCodeSummaryInvalidType    SubCode = "SUMMARY_INVALID_TYPE"

	// Embedding subcodes
	SubCodeEmbeddingNotFound      SubCode = "EMBEDDING_NOT_FOUND"
	SubCodeEmbeddingLLMFailed     SubCode = "EMBEDDING_LLM_FAILED"
	SubCodeEmbeddingNoTranscript   SubCode = "EMBEDDING_NO_TRANSCRIPT"

	// Similarity subcodes
	SubCodeSimilarityNotFound      SubCode = "SIMILARITY_NOT_FOUND"
	SubCodeSimilarityNoEmbeddings  SubCode = "SIMILARITY_NO_EMBEDDINGS"
	SubCodeSimilarityCalculationFailed SubCode = "SIMILARITY_CALCULATION_FAILED"

	// Provider subcodes
	SubCodeProviderNotFound        SubCode = "PROVIDER_NOT_FOUND"
	SubCodeProviderUnavailable     SubCode = "PROVIDER_UNAVAILABLE"
	SubCodeProviderConfigMissing    SubCode = "PROVIDER_CONFIG_MISSING"
	SubCodeProviderRateLimited      SubCode = "PROVIDER_RATE_LIMITED"
	SubCodeProviderQuotaExceeded   SubCode = "PROVIDER_QUOTA_EXCEEDED"
	SubCodeProviderFileTooLarge     SubCode = "PROVIDER_FILE_TOO_LARGE"

	// Database subcodes
	SubCodeDatabaseConnectionFailed SubCode = "DATABASE_CONNECTION_FAILED"
	SubCodeDatabaseQueryFailed      SubCode = "DATABASE_QUERY_FAILED"
	SubCodeDatabaseConstraintViolation SubCode = "DATABASE_CONSTRAINT_VIOLATION"
	SubCodeDatabaseRecordNotFound    SubCode = "DATABASE_RECORD_NOT_FOUND"

	// External service subcodes
	SubCodeYouTubeAPIFailed      SubCode = "YOUTUBE_API_FAILED"
	SubCodeYouTubeDownloadFailed SubCode = "YOUTUBE_DOWNLOAD_FAILED"
	SubCodeLLMAPIFailed          SubCode = "LLM_API_FAILED"
	SubCodeWhisperAPIFailed       SubCode = "WHISPER_API_FAILED"
)

// AppError represents an application error with structured information
type AppError struct {
	Code     ErrorCode `json:"code"`
	SubCode  SubCode   `json:"sub_code"`
	Message  string    `json:"message"`
	Detail   string    `json:"detail,omitempty"`
	HTTPCode int       `json:"-"`
	Err      error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s [%s]: %s - %s", e.Code, e.SubCode, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s [%s]: %s", e.Code, e.SubCode, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code ErrorCode, subCode SubCode, message string) *AppError {
	return &AppError{
		Code:     code,
		SubCode:  subCode,
		Message:  message,
		HTTPCode: getHTTPCode(code),
	}
}

// NewWithDetail creates a new AppError with detail
func NewWithDetail(code ErrorCode, subCode SubCode, message, detail string) *AppError {
	return &AppError{
		Code:     code,
		SubCode:  subCode,
		Message:  message,
		Detail:   detail,
		HTTPCode: getHTTPCode(code),
	}
}

// NewWithError creates a new AppError from an existing error
func NewWithError(code ErrorCode, subCode SubCode, message string, err error) *AppError {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	return &AppError{
		Code:     code,
		SubCode:  subCode,
		Message:  message,
		Detail:   detail,
		HTTPCode: getHTTPCode(code),
		Err:      err,
	}
}

// Wrap wraps an existing error into an AppError
func Wrap(err error, code ErrorCode, subCode SubCode, message string) *AppError {
	if err == nil {
		return nil
	}
	
	// If it's already an AppError, return it
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	
	return NewWithError(code, subCode, message, err)
}

// getHTTPCode maps error codes to HTTP status codes
func getHTTPCode(code ErrorCode) int {
	switch code {
	case ErrorCodeBadRequest, ErrorCodeValidation:
		return http.StatusBadRequest
	case ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrorCodeForbidden:
		return http.StatusForbidden
	case ErrorCodeNotFound, ErrorCodeVideoNotFound, ErrorCodeTranscriptNotFound,
		ErrorCodeSummaryNotFound, ErrorCodeEmbeddingNotFound, ErrorCodeSimilarityNotFound:
		return http.StatusNotFound
	case ErrorCodeConflict, ErrorCodeVideoAlreadyExists:
		return http.StatusConflict
	case ErrorCodeTooLarge, ErrorCodeTranscriptFileTooLarge:
		return http.StatusRequestEntityTooLarge
	case ErrorCodeTimeout:
		return http.StatusRequestTimeout
	case ErrorCodeProviderRateLimit:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors

// ErrVideoNotFound returns a video not found error
func ErrVideoNotFound(videoID string) *AppError {
	return NewWithDetail(
		ErrorCodeVideoNotFound,
		SubCodeVideoNotFound,
		"Video not found",
		fmt.Sprintf("Video with ID %s does not exist", videoID),
	)
}

// ErrVideoInvalidURL returns an invalid video URL error
func ErrVideoInvalidURL(url string) *AppError {
	return NewWithDetail(
		ErrorCodeVideoInvalidURL,
		SubCodeVideoURLInvalid,
		"Invalid video URL",
		fmt.Sprintf("The provided URL '%s' is not a valid YouTube URL", url),
	)
}

// ErrTranscriptNotFound returns a transcript not found error
func ErrTranscriptNotFound(videoID string) *AppError {
	return NewWithDetail(
		ErrorCodeTranscriptNotFound,
		SubCodeTranscriptNotFound,
		"Transcript not found",
		fmt.Sprintf("Transcript for video %s does not exist", videoID),
	)
}

// ErrTranscriptFileTooLarge returns a file too large error
func ErrTranscriptFileTooLarge(fileSize, maxSize int64) *AppError {
	return NewWithDetail(
		ErrorCodeTranscriptFileTooLarge,
		SubCodeTranscriptFileTooLarge,
		"Audio file too large",
		fmt.Sprintf("File size %d bytes exceeds maximum allowed size of %d bytes", fileSize, maxSize),
	)
}

// ErrSummaryNotFound returns a summary not found error
func ErrSummaryNotFound(videoID string) *AppError {
	return NewWithDetail(
		ErrorCodeSummaryNotFound,
		SubCodeSummaryNotFound,
		"Summary not found",
		fmt.Sprintf("Summary for video %s does not exist", videoID),
	)
}

// ErrSummaryNoTranscript returns a no transcript error for summary
func ErrSummaryNoTranscript(videoID string) *AppError {
	return NewWithDetail(
		ErrorCodeSummaryNoTranscript,
		SubCodeSummaryNoTranscript,
		"Transcript required for summary",
		fmt.Sprintf("Video %s does not have a transcript. Please generate a transcript first", videoID),
	)
}

// ErrProviderFileTooLarge returns a provider file too large error
func ErrProviderFileTooLarge(provider string, fileSizeMB, maxSizeMB float64) *AppError {
	return NewWithDetail(
		ErrorCodeProviderUnavailable,
		SubCodeProviderFileTooLarge,
		"File too large for provider",
		fmt.Sprintf("File size %.2f MB exceeds %s provider limit of %.2f MB. Please use local whisper provider", fileSizeMB, provider, maxSizeMB),
	)
}

// ErrDatabaseError returns a database error
func ErrDatabaseError(operation string, err error) *AppError {
	return NewWithError(
		ErrorCodeDatabaseQuery,
		SubCodeDatabaseQueryFailed,
		"Database operation failed",
		err,
	)
}

// ErrInternalError returns an internal server error
func ErrInternalError(message string, err error) *AppError {
	return NewWithError(
		ErrorCodeInternal,
		SubCodeUnknown,
		message,
		err,
	)
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrorCodeNotFound ||
			appErr.Code == ErrorCodeVideoNotFound ||
			appErr.Code == ErrorCodeTranscriptNotFound ||
			appErr.Code == ErrorCodeSummaryNotFound ||
			appErr.Code == ErrorCodeEmbeddingNotFound ||
			appErr.Code == ErrorCodeSimilarityNotFound
	}
	return false
}

// IsBadRequest checks if an error is a bad request error
func IsBadRequest(err error) bool {
	if err == nil {
		return false
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrorCodeBadRequest || appErr.Code == ErrorCodeValidation
	}
	return false
}

