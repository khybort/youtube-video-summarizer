package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/pgvector/pgvector-go"
	"youtube-video-summarizer/backend/internal/models"
)

type EmbeddingRepository interface {
	Save(ctx context.Context, embedding *models.VideoEmbedding) error
	GetByVideoID(ctx context.Context, videoID uuid.UUID, embeddingType string) (*models.VideoEmbedding, error)
	FindSimilar(ctx context.Context, embedding []float32, embeddingType string, limit int, excludeVideoID uuid.UUID) ([]SimilarEmbedding, error)
	CountVideosWithEmbeddings(ctx context.Context, embeddingType string, excludeVideoID uuid.UUID) (int, error)
}

type SimilarEmbedding struct {
	VideoID    uuid.UUID
	Similarity float64
}

type embeddingRepository struct {
	db *gorm.DB
}

func NewEmbeddingRepository(db *gorm.DB) EmbeddingRepository {
	return &embeddingRepository{db: db}
}

func (r *embeddingRepository) Save(ctx context.Context, embedding *models.VideoEmbedding) error {
	if embedding.ID == uuid.Nil {
		embedding.ID = uuid.New()
	}
	
	// Use raw SQL for ON CONFLICT with pgvector
	query := `
		INSERT INTO video_embeddings (id, video_id, embedding_type, embedding, model_used, created_at, updated_at)
		VALUES ($1, $2, $3, $4::vector, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (video_id, embedding_type)
		DO UPDATE SET
			embedding = $4::vector,
			model_used = $5,
			updated_at = CURRENT_TIMESTAMP
	`
	
	// Convert Vector to pgvector for raw SQL
	vec := pgvector.NewVector(embedding.Embedding.Slice())
	
	return r.db.WithContext(ctx).Exec(query,
		embedding.ID, embedding.VideoID, embedding.EmbeddingType,
		vec, embedding.ModelUsed,
	).Error
}

func (r *embeddingRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID, embeddingType string) (*models.VideoEmbedding, error) {
	var embedding models.VideoEmbedding
	err := r.db.WithContext(ctx).
		Where("video_id = ? AND embedding_type = ?", videoID, embeddingType).
		First(&embedding).Error
	if err != nil {
		return nil, err
	}
	return &embedding, nil
}

func (r *embeddingRepository) FindSimilar(
	ctx context.Context,
	embedding []float32,
	embeddingType string,
	limit int,
	excludeVideoID uuid.UUID,
) ([]SimilarEmbedding, error) {
	vec := pgvector.NewVector(embedding)
	
	var results []struct {
		VideoID    uuid.UUID
		Similarity float64
	}

	// Use raw SQL for pgvector similarity search
	query := `
		SELECT
			video_id,
			1 - (embedding <=> $1::vector) as similarity
		FROM video_embeddings
		WHERE embedding_type = $2
		AND video_id != $3
		ORDER BY embedding <=> $1::vector
		LIMIT $4
	`

	err := r.db.WithContext(ctx).
		Raw(query, vec, embeddingType, excludeVideoID, limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	similarEmbeddings := make([]SimilarEmbedding, len(results))
	for i, r := range results {
		similarEmbeddings[i] = SimilarEmbedding{
			VideoID:    r.VideoID,
			Similarity: r.Similarity,
		}
	}

	return similarEmbeddings, nil
}

func (r *embeddingRepository) CountVideosWithEmbeddings(ctx context.Context, embeddingType string, excludeVideoID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.VideoEmbedding{}).
		Where("embedding_type = ? AND video_id != ?", embeddingType, excludeVideoID).
		Count(&count).Error
	return int(count), err
}
