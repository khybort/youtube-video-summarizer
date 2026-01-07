package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Video struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	YouTubeID    string    `gorm:"type:varchar(255);uniqueIndex;not null;column:youtube_id" json:"youtube_id"`
	Title        string    `gorm:"type:text;not null" json:"title"`
	Description  string    `gorm:"type:text" json:"description"`
	ChannelID    string    `gorm:"type:varchar(255);not null;index" json:"channel_id"`
	ChannelName  string    `gorm:"type:varchar(255);not null" json:"channel_name"`
	Duration     int       `gorm:"not null" json:"duration"` // seconds
	ViewCount    int64     `gorm:"default:0" json:"view_count"`
	LikeCount    int64     `gorm:"default:0" json:"like_count"`
	PublishedAt  time.Time `gorm:"not null" json:"published_at"`
	ThumbnailURL string    `gorm:"type:text" json:"thumbnail_url"`
	Tags         pq.StringArray `gorm:"type:text[];default:'{}'" json:"tags"`
	Category     string    `gorm:"type:varchar(100)" json:"category"`
	Status       string    `gorm:"type:varchar(50);default:'pending';index" json:"status"` // pending, processing, completed, error
	HasTranscript bool     `gorm:"default:false" json:"has_transcript"`
	HasSummary   bool     `gorm:"default:false" json:"has_summary"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index:idx_videos_created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Video) TableName() string {
	return "videos"
}

type Transcript struct {
	ID        uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID   uuid.UUID          `gorm:"type:uuid;not null;index" json:"video_id"`
	Language  string             `gorm:"type:varchar(10);not null" json:"language"`
	Source    string             `gorm:"type:varchar(50);not null" json:"source"` // youtube, whisper
	Content   string             `gorm:"type:text;not null" json:"content"`
	Segments  TranscriptSegments `gorm:"type:jsonb" json:"segments"`
	CreatedAt time.Time          `gorm:"autoCreateTime" json:"created_at"`
	Video     Video              `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Transcript) TableName() string {
	return "transcripts"
}

type TranscriptSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// TranscriptSegments is a custom type for JSONB serialization
type TranscriptSegments []TranscriptSegment

// Value implements driver.Valuer interface for JSONB
func (ts TranscriptSegments) Value() (driver.Value, error) {
	if len(ts) == 0 {
		return "[]", nil
	}
	return json.Marshal(ts)
}

// Scan implements sql.Scanner interface for JSONB
func (ts *TranscriptSegments) Scan(value interface{}) error {
	if value == nil {
		*ts = TranscriptSegments{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	if len(bytes) == 0 {
		*ts = TranscriptSegments{}
		return nil
	}

	return json.Unmarshal(bytes, ts)
}

type Summary struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID     uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	ModelUsed   string    `gorm:"type:varchar(100);not null" json:"model_used"`
	SummaryType string    `gorm:"type:varchar(50);not null" json:"summary_type"` // short, detailed, bullet_points
	Content     string    `gorm:"type:text;not null" json:"content"`
	KeyPoints   pq.StringArray `gorm:"type:text[];default:'{}'" json:"key_points"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	Video       Video     `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Summary) TableName() string {
	return "summaries"
}

type VideoEmbedding struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID       uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	EmbeddingType string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_embeddings_video_type" json:"embedding_type"` // title, description, transcript, combined
	Embedding     Vector    `gorm:"type:vector(768);not null" json:"embedding"`
	ModelUsed     string    `gorm:"type:varchar(100);not null" json:"model_used"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Video         Video     `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE" json:"-"`
}

func (VideoEmbedding) TableName() string {
	return "video_embeddings"
}

type VideoSimilarity struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID1            uuid.UUID `gorm:"type:uuid;not null;index;column:video_id_1" json:"video_id_1"`
	VideoID2            uuid.UUID `gorm:"type:uuid;not null;index;column:video_id_2" json:"video_id_2"`
	TitleSimilarity     float64   `gorm:"type:float" json:"title_similarity"`
	DescSimilarity      float64   `gorm:"type:float" json:"desc_similarity"`
	TranscriptSimilarity float64   `gorm:"type:float" json:"transcript_similarity"`
	CombinedSimilarity  float64   `gorm:"type:float;not null;index" json:"combined_similarity"`
	ComparisonType      string    `gorm:"type:varchar(50);default:'auto'" json:"comparison_type"`
	ModelUsed           string    `gorm:"type:varchar(100)" json:"model_used"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	Video1              Video     `gorm:"foreignKey:VideoID1;constraint:OnDelete:CASCADE" json:"-"`
	Video2              Video     `gorm:"foreignKey:VideoID2;constraint:OnDelete:CASCADE" json:"-"`
}

func (VideoSimilarity) TableName() string {
	return "video_similarities"
}

// BeforeCreate hook to ensure video_id_1 < video_id_2
func (vs *VideoSimilarity) BeforeCreate(tx *gorm.DB) error {
	if vs.VideoID1.String() > vs.VideoID2.String() {
		vs.VideoID1, vs.VideoID2 = vs.VideoID2, vs.VideoID1
	}
	return nil
}

type SimilarityResult struct {
	VideoID1            uuid.UUID
	VideoID2            uuid.UUID
	TitleSimilarity     float64
	DescSimilarity      float64
	TranscriptSimilarity float64
	CombinedSimilarity  float64
}

type SimilarVideo struct {
	Video           *Video
	SimilarityScore float64
	ComparisonType  string
}
