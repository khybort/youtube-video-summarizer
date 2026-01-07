package embedding

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/internal/services/cost"
	"youtube-video-summarizer/backend/internal/services/provider"
	"youtube-video-summarizer/backend/internal/services/transcript"
	"youtube-video-summarizer/backend/pkg/llm"
)

type Service struct {
	embeddingRepo     repository.EmbeddingRepository
	providerFactory   *provider.ProviderFactory
	costService       *cost.Service
	transcriptService *transcript.Service
	logger            *zap.Logger
}

func NewService(
	embeddingRepo repository.EmbeddingRepository,
	providerFactory *provider.ProviderFactory,
	costService *cost.Service,
	transcriptService *transcript.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		embeddingRepo:     embeddingRepo,
		providerFactory:   providerFactory,
		costService:       costService,
		transcriptService: transcriptService,
		logger:            logger,
	}
}

type VideoEmbeddings struct {
	VideoID             uuid.UUID
	TitleEmbedding      []float32
	DescriptionEmbedding []float32
	TranscriptEmbedding []float32
	CombinedEmbedding   []float32
}

func (s *Service) GenerateVideoEmbeddings(ctx context.Context, video *models.Video, transcript string) (*VideoEmbeddings, error) {
	// Get LLM provider from settings
	llmProvider, err := s.providerFactory.GetLLMProvider(ctx, "embedding")
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider: %w", err)
	}

	// If transcriptContent is empty, try to fetch it
	if transcript == "" && s.transcriptService != nil {
		t, err := s.transcriptService.GetByVideoID(ctx, video.ID)
		if err == nil && t != nil {
			transcript = t.Content
			s.logger.Debug("Fetched transcript for embedding generation", zap.String("video_id", video.ID.String()))
		} else {
			s.logger.Warn("Transcript not found for embedding generation, proceeding without it", zap.String("video_id", video.ID.String()), zap.Error(err))
		}
	}

	// Generate title embedding
	titleEmb, err := llmProvider.GenerateEmbedding(ctx, video.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to generate title embedding: %w", err)
	}

	// Generate description embedding
	descEmb, err := llmProvider.GenerateEmbedding(ctx, video.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to generate description embedding: %w", err)
	}

	// Generate transcript embedding (chunked for long transcripts)
	var transcriptEmb []float32
	if transcript != "" {
		transcriptEmb, err = s.generateTranscriptEmbedding(ctx, transcript, llmProvider)
		if err != nil {
			return nil, fmt.Errorf("failed to generate transcript embedding: %w", err)
		}
	}

	// Generate combined embedding (weighted average)
	combinedEmb := s.combineEmbeddings(titleEmb, descEmb, transcriptEmb)

	embeddings := &VideoEmbeddings{
		VideoID:              video.ID,
		TitleEmbedding:       titleEmb,
		DescriptionEmbedding: descEmb,
		TranscriptEmbedding:  transcriptEmb,
		CombinedEmbedding:    combinedEmb,
	}

	// Save embeddings
	if err := s.saveEmbeddings(ctx, embeddings, video.ID, llmProvider); err != nil {
		return nil, fmt.Errorf("failed to save embeddings: %w", err)
	}

	// Record token usage for embeddings
	if s.costService != nil {
		modelInfo := llmProvider.GetModelInfo()
		// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
		estimatedTokens := (len(video.Title) + len(video.Description) + len(transcript)) / 4
		
		_ = s.costService.RecordUsage(
			ctx,
			video.ID,
			"embedding",
			modelInfo.Provider,
			modelInfo.Name,
			estimatedTokens,
			0, // Embeddings don't have output tokens
		)
	}

	return embeddings, nil
}

func (s *Service) generateTranscriptEmbedding(ctx context.Context, transcript string, llmProvider llm.LLMProvider) ([]float32, error) {
	// Chunk long transcripts
	chunks := s.chunkText(transcript, 512, 50)

	if len(chunks) == 0 {
		return nil, fmt.Errorf("empty transcript")
	}

	// Generate embeddings for each chunk
	chunkEmbeddings, err := llmProvider.GenerateBatchEmbeddings(ctx, chunks)
	if err != nil {
		return nil, err
	}

	// Mean pooling
	return s.meanPool(chunkEmbeddings), nil
}

func (s *Service) chunkText(text string, chunkSize, overlap int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var chunks []string
	var currentChunk []string
	currentSize := 0

	for _, word := range words {
		currentChunk = append(currentChunk, word)
		currentSize++

		if currentSize >= chunkSize {
			chunks = append(chunks, strings.Join(currentChunk, " "))
			// Keep overlap words for next chunk
			if overlap > 0 && len(currentChunk) > overlap {
				currentChunk = currentChunk[len(currentChunk)-overlap:]
				currentSize = overlap
			} else {
				currentChunk = []string{}
				currentSize = 0
			}
		}
	}

	// Add remaining chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

func (s *Service) meanPool(embeddings [][]float32) []float32 {
	if len(embeddings) == 0 {
		return nil
	}

	dim := len(embeddings[0])
	result := make([]float32, dim)

	for _, emb := range embeddings {
		for i, v := range emb {
			result[i] += v
		}
	}

	for i := range result {
		result[i] /= float32(len(embeddings))
	}

	return result
}

func (s *Service) combineEmbeddings(title, desc, transcript []float32) []float32 {
	// Weighted combination: title 15%, description 20%, transcript 65%
	weights := map[string]float32{
		"title":      0.15,
		"description": 0.20,
		"transcript":  0.65,
	}

	// Use available embeddings
	var embeddings [][]float32
	var totalWeight float32

	if len(title) > 0 {
		embeddings = append(embeddings, title)
		totalWeight += weights["title"]
	}
	if len(desc) > 0 {
		embeddings = append(embeddings, desc)
		totalWeight += weights["description"]
	}
	if len(transcript) > 0 {
		embeddings = append(embeddings, transcript)
		totalWeight += weights["transcript"]
	}

	if len(embeddings) == 0 {
		return nil
	}

	// Weighted average
	dim := len(embeddings[0])
	result := make([]float32, dim)

	weightIdx := 0
	for _, emb := range embeddings {
		weight := weights[map[int]string{0: "title", 1: "description", 2: "transcript"}[weightIdx]]
		for i, v := range emb {
			result[i] += v * weight
		}
		weightIdx++
	}

	// Normalize
	if totalWeight > 0 {
		for i := range result {
			result[i] /= totalWeight
		}
	}

	return result
}

func (s *Service) saveEmbeddings(ctx context.Context, embeddings *VideoEmbeddings, videoID uuid.UUID, llmProvider llm.LLMProvider) error {
	modelName := llmProvider.GetModelInfo().Name

	// Save title embedding
	if len(embeddings.TitleEmbedding) > 0 {
		if err := s.embeddingRepo.Save(ctx, &models.VideoEmbedding{
			VideoID:       videoID,
			EmbeddingType: "title",
			Embedding:     models.Vector{Data: embeddings.TitleEmbedding},
			ModelUsed:     modelName,
		}); err != nil {
			return err
		}
	}

	// Save description embedding
	if len(embeddings.DescriptionEmbedding) > 0 {
		if err := s.embeddingRepo.Save(ctx, &models.VideoEmbedding{
			VideoID:       videoID,
			EmbeddingType: "description",
			Embedding:     models.Vector{Data: embeddings.DescriptionEmbedding},
			ModelUsed:     modelName,
		}); err != nil {
			return err
		}
	}

	// Save transcript embedding
	if len(embeddings.TranscriptEmbedding) > 0 {
		if err := s.embeddingRepo.Save(ctx, &models.VideoEmbedding{
			VideoID:       videoID,
			EmbeddingType: "transcript",
			Embedding:     models.Vector{Data: embeddings.TranscriptEmbedding},
			ModelUsed:     modelName,
		}); err != nil {
			return err
		}
	}

	// Save combined embedding
	if len(embeddings.CombinedEmbedding) > 0 {
		if err := s.embeddingRepo.Save(ctx, &models.VideoEmbedding{
			VideoID:       videoID,
			EmbeddingType: "combined",
			Embedding:     models.Vector{Data: embeddings.CombinedEmbedding},
			ModelUsed:     modelName,
		}); err != nil {
			return err
		}
	}

	return nil
}

