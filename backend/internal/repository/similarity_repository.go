package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
)

type SimilarityRepository interface {
	Save(ctx context.Context, result *models.SimilarityResult) error
	GetByVideoPair(ctx context.Context, videoID1, videoID2 uuid.UUID) (*models.SimilarityResult, error)
	GetSimilarVideos(ctx context.Context, videoID uuid.UUID, limit int, minScore float64) ([]models.SimilarVideo, error)
}

type similarityRepository struct {
	db *gorm.DB
}

func NewSimilarityRepository(db *gorm.DB) SimilarityRepository {
	return &similarityRepository{db: db}
}

func (r *similarityRepository) Save(ctx context.Context, result *models.SimilarityResult) error {
	// Ensure consistent ordering (smaller ID first)
	id1, id2 := result.VideoID1, result.VideoID2
	if id1.String() > id2.String() {
		id1, id2 = id2, id1
	}

	// similarity variable is not needed, using raw SQL

	// Use raw SQL for unique constraint handling
	query := `
		INSERT INTO video_similarities (
			video_id_1, video_id_2,
			title_similarity, desc_similarity, transcript_similarity,
			combined_similarity, comparison_type
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (LEAST(video_id_1, video_id_2), GREATEST(video_id_1, video_id_2))
		DO UPDATE SET
			title_similarity = $3,
			desc_similarity = $4,
			transcript_similarity = $5,
			combined_similarity = $6,
			comparison_type = $7
	`

	return r.db.WithContext(ctx).Exec(query,
		id1, id2,
		result.TitleSimilarity, result.DescSimilarity, result.TranscriptSimilarity,
		result.CombinedSimilarity, "auto",
	).Error
}

func (r *similarityRepository) GetByVideoPair(ctx context.Context, videoID1, videoID2 uuid.UUID) (*models.SimilarityResult, error) {
	var similarity models.VideoSimilarity
	err := r.db.WithContext(ctx).
		Where("(video_id_1 = ? AND video_id_2 = ?) OR (video_id_1 = ? AND video_id_2 = ?)",
			videoID1, videoID2, videoID2, videoID1).
		First(&similarity).Error

	if err != nil {
		return nil, err
	}

	return &models.SimilarityResult{
		VideoID1:            similarity.VideoID1,
		VideoID2:            similarity.VideoID2,
		TitleSimilarity:     similarity.TitleSimilarity,
		DescSimilarity:      similarity.DescSimilarity,
		TranscriptSimilarity: similarity.TranscriptSimilarity,
		CombinedSimilarity:  similarity.CombinedSimilarity,
	}, nil
}

func (r *similarityRepository) GetSimilarVideos(ctx context.Context, videoID uuid.UUID, limit int, minScore float64) ([]models.SimilarVideo, error) {
	var results []struct {
		SimilarVideoID uuid.UUID
		Similarity     float64
	}

	query := `
		SELECT 
			CASE WHEN video_id_1 = ? THEN video_id_2 ELSE video_id_1 END as similar_video_id,
			combined_similarity
		FROM video_similarities
		WHERE (video_id_1 = ? OR video_id_2 = ?)
		AND combined_similarity >= ?
		ORDER BY combined_similarity DESC
		LIMIT ?
	`

	err := r.db.WithContext(ctx).
		Raw(query, videoID, videoID, videoID, minScore, limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	similarVideos := make([]models.SimilarVideo, len(results))
	for i, r := range results {
		similarVideos[i] = models.SimilarVideo{
			Video: &models.Video{
				ID: r.SimilarVideoID,
			},
			SimilarityScore: r.Similarity,
			ComparisonType:  "combined",
		}
	}

	return similarVideos, nil
}
