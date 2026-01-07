package similarity

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/pkg/errors"
	"youtube-video-summarizer/backend/pkg/youtube"
)

type Service struct {
	similarityRepo repository.SimilarityRepository
	embeddingRepo  repository.EmbeddingRepository
	videoRepo      repository.VideoRepository
	youtubeClient  *youtube.Client
	logger         *zap.Logger
}

func NewService(
	similarityRepo repository.SimilarityRepository,
	embeddingRepo repository.EmbeddingRepository,
	videoRepo repository.VideoRepository,
	youtubeClient *youtube.Client,
	logger *zap.Logger,
) *Service {
	return &Service{
		similarityRepo: similarityRepo,
		embeddingRepo:  embeddingRepo,
		videoRepo:      videoRepo,
		youtubeClient:  youtubeClient,
		logger:         logger,
	}
}

func (s *Service) CalculateSimilarity(ctx context.Context, videoID1, videoID2 uuid.UUID) (*models.SimilarityResult, error) {
	// Get embeddings for both videos
	emb1, err := s.embeddingRepo.GetByVideoID(ctx, videoID1, "combined")
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings for video 1: %w", err)
	}

	emb2, err := s.embeddingRepo.GetByVideoID(ctx, videoID2, "combined")
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings for video 2: %w", err)
	}

	// Calculate cosine similarity for combined embeddings
	combinedSim := cosineSimilarity(emb1.Embedding.Slice(), emb2.Embedding.Slice())

	// Get individual similarities
	titleSim := s.getSimilarityByType(ctx, videoID1, videoID2, "title")
	descSim := s.getSimilarityByType(ctx, videoID1, videoID2, "description")
	transcriptSim := s.getSimilarityByType(ctx, videoID1, videoID2, "transcript")

	result := &models.SimilarityResult{
		VideoID1:            videoID1,
		VideoID2:            videoID2,
		TitleSimilarity:     titleSim,
		DescSimilarity:      descSim,
		TranscriptSimilarity: transcriptSim,
		CombinedSimilarity:  combinedSim,
	}

	// Save similarity
	if err := s.similarityRepo.Save(ctx, result); err != nil {
		s.logger.Warn("Failed to save similarity", zap.Error(err))
	}

	return result, nil
}

func (s *Service) getSimilarityByType(ctx context.Context, videoID1, videoID2 uuid.UUID, embeddingType string) float64 {
	emb1, err := s.embeddingRepo.GetByVideoID(ctx, videoID1, embeddingType)
	if err != nil {
		return 0
	}

	emb2, err := s.embeddingRepo.GetByVideoID(ctx, videoID2, embeddingType)
	if err != nil {
		return 0
	}

	return cosineSimilarity(emb1.Embedding.Slice(), emb2.Embedding.Slice())
}

func (s *Service) FindSimilarVideos(ctx context.Context, videoID uuid.UUID, limit int, minThreshold float64) ([]models.SimilarVideo, error) {
	// Get video to get YouTube ID
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrVideoNotFound(videoID.String())
		}
		return nil, errors.Wrap(err, errors.ErrorCodeVideoNotFound, errors.SubCodeVideoNotFound, "Failed to get video")
	}

	// Fetch similar videos directly from YouTube
	if s.youtubeClient == nil {
		return nil, errors.New(
			errors.ErrorCodeYouTubeAPI,
			errors.SubCodeYouTubeAPIFailed,
			"YouTube client not configured",
		)
	}

	youtubeVideos, err := s.fetchSimilarFromYouTube(ctx, video.YouTubeID, limit)
	if err != nil {
		return nil, err
	}

	return youtubeVideos, nil
}

// fetchSimilarFromYouTube fetches similar videos from YouTube API
func (s *Service) fetchSimilarFromYouTube(ctx context.Context, youtubeID string, limit int) ([]models.SimilarVideo, error) {
	if limit <= 0 {
		return []models.SimilarVideo{}, nil
	}

	s.logger.Info("Fetching similar videos from YouTube",
		zap.String("youtube_id", youtubeID),
		zap.Int("limit", limit))

	youtubeVideos, err := s.youtubeClient.SearchRelatedVideos(ctx, youtubeID, limit)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorCodeYouTubeAPI, errors.SubCodeYouTubeAPIFailed, "Failed to fetch related videos from YouTube")
	}

	// Convert YouTube VideoInfo to SimilarVideo
	results := make([]models.SimilarVideo, 0, len(youtubeVideos))
	for _, ytVideo := range youtubeVideos {
		// Check if video already exists in database
		existingVideo, err := s.videoRepo.GetByYouTubeID(ctx, ytVideo.ID)
		if err == nil && existingVideo != nil {
			// Video exists in database, use it
			results = append(results, models.SimilarVideo{
				Video:           existingVideo,
				SimilarityScore: 0.0, // YouTube relevance, not calculated similarity
				ComparisonType:  "youtube",
			})
		} else {
			// Create a minimal video object from YouTube data
			video := &models.Video{
				YouTubeID:    ytVideo.ID,
				Title:        ytVideo.Title,
				Description:  ytVideo.Description,
				ChannelID:    ytVideo.ChannelID,
				ChannelName:  ytVideo.ChannelName,
				Duration:     ytVideo.Duration,
				ViewCount:    ytVideo.ViewCount,
				LikeCount:    ytVideo.LikeCount,
				PublishedAt:  ytVideo.PublishedAt,
				ThumbnailURL: ytVideo.ThumbnailURL,
				Tags:         ytVideo.Tags,
				Category:     ytVideo.Category,
				Status:       "pending", // Not in database yet
			}
			results = append(results, models.SimilarVideo{
				Video:           video,
				SimilarityScore: 0.0, // YouTube relevance, not calculated similarity
				ComparisonType:  "youtube",
			})
		}
	}

	s.logger.Info("Fetched similar videos from YouTube",
		zap.String("youtube_id", youtubeID),
		zap.Int("count", len(results)))

	return results, nil
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64

	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

