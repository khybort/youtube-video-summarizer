package kafka

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a base event structure
type Event struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	Timestamp   time.Time `json:"timestamp"`
	VideoID     string    `json:"video_id"`
	YouTubeID   string    `json:"youtube_id,omitempty"`
}

// VideoCreatedEvent is published when a new video is created
type VideoCreatedEvent struct {
	Event
	Title        string    `json:"title"`
	ChannelID    string    `json:"channel_id"`
	ChannelName  string    `json:"channel_name"`
	Duration     int       `json:"duration"`
	PublishedAt  time.Time `json:"published_at"`
	ThumbnailURL string    `json:"thumbnail_url"`
}

// NewVideoCreatedEvent creates a new VideoCreatedEvent
func NewVideoCreatedEvent(videoID, youtubeID, title, channelID, channelName string, duration int, publishedAt time.Time, thumbnailURL string) *VideoCreatedEvent {
	return &VideoCreatedEvent{
		Event: Event{
			EventID:   uuid.New().String(),
			EventType: "video.created",
			Timestamp: time.Now(),
			VideoID:   videoID,
			YouTubeID: youtubeID,
		},
		Title:        title,
		ChannelID:    channelID,
		ChannelName:  channelName,
		Duration:     duration,
		PublishedAt:  publishedAt,
		ThumbnailURL: thumbnailURL,
	}
}

// TranscriptRequestedEvent is published when transcript generation is requested
type TranscriptRequestedEvent struct {
	Event
	Priority int `json:"priority"` // Higher priority = process first
}

// NewTranscriptRequestedEvent creates a new TranscriptRequestedEvent
func NewTranscriptRequestedEvent(videoID, youtubeID string, priority int) *TranscriptRequestedEvent {
	return &TranscriptRequestedEvent{
		Event: Event{
			EventID:   uuid.New().String(),
			EventType: "transcript.requested",
			Timestamp: time.Now(),
			VideoID:   videoID,
			YouTubeID: youtubeID,
		},
		Priority: priority,
	}
}

// EmbeddingRequestedEvent is published when embedding generation is requested
type EmbeddingRequestedEvent struct {
	Event
	TranscriptContent string `json:"transcript_content,omitempty"`
	Priority          int    `json:"priority"`
}

// NewEmbeddingRequestedEvent creates a new EmbeddingRequestedEvent
func NewEmbeddingRequestedEvent(videoID, youtubeID, transcriptContent string, priority int) *EmbeddingRequestedEvent {
	return &EmbeddingRequestedEvent{
		Event: Event{
			EventID:   uuid.New().String(),
			EventType: "embedding.requested",
			Timestamp: time.Now(),
			VideoID:   videoID,
			YouTubeID: youtubeID,
		},
		TranscriptContent: transcriptContent,
		Priority:          priority,
	}
}

// SimilarityRequestedEvent is published when similarity calculation is requested
type SimilarityRequestedEvent struct {
	Event
	TargetVideoID string `json:"target_video_id,omitempty"` // If empty, calculate with all videos
	Priority        int    `json:"priority"`
}

// NewSimilarityRequestedEvent creates a new SimilarityRequestedEvent
func NewSimilarityRequestedEvent(videoID, youtubeID string, targetVideoID string, priority int) *SimilarityRequestedEvent {
	return &SimilarityRequestedEvent{
		Event: Event{
			EventID:   uuid.New().String(),
			EventType: "similarity.requested",
			Timestamp: time.Now(),
			VideoID:   videoID,
			YouTubeID: youtubeID,
		},
		TargetVideoID: targetVideoID,
		Priority:      priority,
	}
}

// AnalysisCompletedEvent is published when video analysis is completed
type AnalysisCompletedEvent struct {
	Event
	HasTranscript bool `json:"has_transcript"`
	HasSummary    bool `json:"has_summary"`
	HasEmbedding  bool `json:"has_embedding"`
	Duration      int  `json:"duration_seconds"` // Processing duration
}

// NewAnalysisCompletedEvent creates a new AnalysisCompletedEvent
func NewAnalysisCompletedEvent(videoID, youtubeID string, hasTranscript, hasSummary, hasEmbedding bool, duration int) *AnalysisCompletedEvent {
	return &AnalysisCompletedEvent{
		Event: Event{
			EventID:   uuid.New().String(),
			EventType: "analysis.completed",
			Timestamp: time.Now(),
			VideoID:   videoID,
			YouTubeID: youtubeID,
		},
		HasTranscript: hasTranscript,
		HasSummary:    hasSummary,
		HasEmbedding:  hasEmbedding,
		Duration:      duration,
	}
}

// AnalysisFailedEvent is published when video analysis fails
type AnalysisFailedEvent struct {
	Event
	Error     string `json:"error"`
	Stage     string `json:"stage"` // "transcript", "embedding", "similarity", etc.
	Retryable bool   `json:"retryable"`
}

// NewAnalysisFailedEvent creates a new AnalysisFailedEvent
func NewAnalysisFailedEvent(videoID, youtubeID, stage, errorMsg string, retryable bool) *AnalysisFailedEvent {
	return &AnalysisFailedEvent{
		Event: Event{
			EventID:   uuid.New().String(),
			EventType: "analysis.failed",
			Timestamp: time.Now(),
			VideoID:   videoID,
			YouTubeID: youtubeID,
		},
		Error:     errorMsg,
		Stage:     stage,
		Retryable: retryable,
	}
}

