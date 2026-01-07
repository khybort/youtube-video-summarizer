package handlers

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
	kafkaservice "youtube-video-summarizer/backend/internal/services/kafka"
	"youtube-video-summarizer/backend/pkg/errors"
)

func RegisterVideoRoutes(
	router *gin.RouterGroup,
	videoService VideoService,
	transcriptService TranscriptService,
	summaryService SummaryService,
	embeddingService EmbeddingService,
	similarityService SimilarityService,
	videoEventService *kafkaservice.VideoEventService,
	logger *zap.Logger,
) {
	handler := &VideoHandler{
		videoService:      videoService,
		transcriptService: transcriptService,
		summaryService:    summaryService,
		embeddingService:  embeddingService,
		similarityService: similarityService,
		videoEventService: videoEventService,
		logger:            logger,
	}
	
	videos := router.Group("/videos")
	{
		videos.POST("", handler.CreateVideo)
		videos.GET("", handler.ListVideos)
		videos.GET("/:id", handler.GetVideo)
		videos.DELETE("/:id", handler.DeleteVideo)
		videos.POST("/:id/analyze", handler.AnalyzeVideo)
		videos.GET("/:id/transcript", handler.GetTranscript)
	videos.GET("/:id/transcript/languages", handler.GetAvailableLanguages)
		videos.GET("/:id/summary", handler.GetSummary)
		videos.POST("/:id/summarize", handler.SummarizeVideo)
		videos.GET("/:id/similar", handler.GetSimilarVideos)
	}
}

type VideoHandler struct {
	videoService      VideoService
	transcriptService TranscriptService
	summaryService    SummaryService
	embeddingService  EmbeddingService
	similarityService SimilarityService
	videoEventService *kafkaservice.VideoEventService
	logger            *zap.Logger
}

func (h *VideoHandler) CreateVideo(c *gin.Context) {
	var req struct {
		URL      string `json:"url"`
		YouTubeID string `json:"youtube_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid request body",
		))
		return
	}

	// Support both url and youtube_id fields
	videoURL := req.URL
	if videoURL == "" {
		videoURL = req.YouTubeID
	}
	if videoURL == "" {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeMissingParameter,
			"url or youtube_id is required",
		))
		return
	}

	video, err := h.videoService.CreateFromURL(c.Request.Context(), videoURL)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	// Publish video.created event to Kafka
	if h.videoEventService != nil {
		if err := h.videoEventService.PublishVideoCreated(c.Request.Context(), video); err != nil {
			h.logger.Warn("Failed to publish video.created event, continuing with direct processing",
				zap.String("video_id", video.ID.String()),
				zap.Error(err),
			)
			// Fallback to direct processing if Kafka is unavailable
			go h.performAnalysisDirect(context.Background(), video.ID)
		} else {
			// Publish transcript request event
			if err := h.videoEventService.PublishTranscriptRequested(
				c.Request.Context(),
				video.ID,
				video.YouTubeID,
				1, // Normal priority
			); err != nil {
				h.logger.Warn("Failed to publish transcript.requested event",
					zap.String("video_id", video.ID.String()),
					zap.Error(err),
				)
			}
		}
	} else {
		// Kafka not enabled, use direct processing
		go h.performAnalysisDirect(context.Background(), video.ID)
	}

	h.logger.Info("Video created", zap.String("video_id", video.ID.String()))
	c.JSON(http.StatusOK, video)
}

func (h *VideoHandler) ListVideos(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	videos, total, err := h.videoService.List(c.Request.Context(), limit, offset)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videos": videos,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *VideoHandler) GetVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid video ID format",
		))
		return
	}

	video, err := h.videoService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			errors.AbortWithError(c, errors.ErrVideoNotFound(id.String()))
		} else {
			errors.HandleError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, video)
}

func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid video ID format",
		))
		return
	}

	if err := h.videoService.Delete(c.Request.Context(), id); err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video deleted"})
}

func (h *VideoHandler) AnalyzeVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid video ID format",
		))
		return
	}

	// Update status to processing
	if err := h.videoService.UpdateStatus(c.Request.Context(), id, "processing"); err != nil {
		errors.HandleError(c, err)
		return
	}

	// Get video to get YouTube ID
	video, err := h.videoService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			errors.AbortWithError(c, errors.ErrVideoNotFound(id.String()))
		} else {
			errors.HandleError(c, err)
		}
		return
	}

	// Publish transcript request event
	if h.videoEventService != nil {
		if err := h.videoEventService.PublishTranscriptRequested(
			c.Request.Context(),
			id,
			video.YouTubeID,
			1, // Normal priority
		); err != nil {
			h.logger.Warn("Failed to publish transcript.requested event, using direct processing",
				zap.String("video_id", id.String()),
				zap.Error(err),
			)
			go h.performAnalysisDirect(context.Background(), id)
		}
	} else {
		go h.performAnalysisDirect(context.Background(), id)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Analysis started",
		"video_id": id.String(),
		"status":   "processing",
	})
}

// performAnalysisDirect performs analysis directly (fallback when Kafka is unavailable)
func (h *VideoHandler) performAnalysisDirect(ctx context.Context, videoID uuid.UUID) {
	// 1. Get or create transcript
	transcript, err := h.transcriptService.GetOrCreateTranscript(ctx, videoID)
	if err != nil {
		h.logger.Error("Failed to get transcript", zap.String("video_id", videoID.String()), zap.Error(err))
		h.videoService.UpdateStatus(ctx, videoID, "error")
		if h.videoEventService != nil {
			video, _ := h.videoService.GetByID(ctx, videoID)
			if video != nil {
				_ = h.videoEventService.PublishAnalysisFailed(ctx, videoID, video.YouTubeID, "transcript", err.Error(), true)
			}
		}
		return
	}

	// 2. Get video
	video, err := h.videoService.GetByID(ctx, videoID)
	if err != nil {
		h.logger.Error("Failed to get video", zap.String("video_id", videoID.String()), zap.Error(err))
		return
	}

	// 3. Generate embeddings
	_, err = h.embeddingService.GenerateVideoEmbeddings(ctx, video, transcript.Content)
	if err != nil {
		h.logger.Error("Failed to generate embeddings", zap.String("video_id", videoID.String()), zap.Error(err))
		h.videoService.UpdateStatus(ctx, videoID, "error")
		if h.videoEventService != nil {
			_ = h.videoEventService.PublishAnalysisFailed(ctx, videoID, video.YouTubeID, "embedding", err.Error(), true)
		}
		return
	}

	// 4. Calculate similarities with existing videos (async, non-blocking)
	go h.calculateSimilarities(ctx, videoID)

	// 5. Update status
	h.videoService.UpdateStatus(ctx, videoID, "completed")
	h.logger.Info("Video analysis completed", zap.String("video_id", videoID.String()))
	
	if h.videoEventService != nil {
		_ = h.videoEventService.PublishAnalysisCompleted(ctx, videoID, video.YouTubeID, true, false, true, 0)
	}
}

func (h *VideoHandler) calculateSimilarities(ctx context.Context, videoID uuid.UUID) {
	// Get all other videos and calculate similarities
	// This is a simplified version - in production, use batch processing
	// For now, similarity calculation happens on-demand when GetSimilarVideos is called
	h.logger.Info("Similarity calculation queued", zap.String("video_id", videoID.String()))
}

func (h *VideoHandler) GetTranscript(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid video ID format",
		))
		return
	}

	// Get language parameter from query string
	languageCode := c.Query("language")

	// First try to get existing transcript
	transcript, err := h.transcriptService.GetByVideoID(c.Request.Context(), id)
	if err == nil && transcript != nil {
		// If language is specified and matches existing, return it
		if languageCode == "" || transcript.Language == languageCode {
			c.JSON(http.StatusOK, transcript)
			return
		}
		// Language doesn't match, will create new one below
	}

	// If not found or language mismatch, try to create it (this may take time for large videos)
	// Use a longer timeout context for transcript generation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	// If language is specified, use it; otherwise use existing transcript service logic
	if languageCode != "" {
		// Create transcript with specific language
		transcript, err = h.transcriptService.GetOrCreateTranscript(ctx, id, languageCode)
	} else {
		// Use default behavior (will try to get existing or create with default language)
		transcript, err = h.transcriptService.GetOrCreateTranscript(ctx, id)
	}
	
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, transcript)
}

func (h *VideoHandler) GetAvailableLanguages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid video ID format",
		))
		return
	}

	// Get video to access YouTube ID
	video, err := h.videoService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			errors.AbortWithError(c, errors.ErrVideoNotFound(id.String()))
		} else {
			errors.HandleError(c, err)
		}
		return
	}

	// List available languages
	languages, err := h.transcriptService.ListAvailableLanguages(c.Request.Context(), video.YouTubeID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"video_id": id.String(),
		"languages": languages,
	})
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func (h *VideoHandler) GetSummary(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid video ID format",
		))
		return
	}

	// Get language from query parameter, default to "auto"
	language := c.Query("language")
	if language == "" {
		language = "auto"
	}

	// First try to get existing summary
	summary, err := h.summaryService.GetByVideoID(c.Request.Context(), id)
	if err == nil && summary != nil {
		c.JSON(http.StatusOK, summary)
		return
	}

	// If not found, try to create it
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	// Get video to access YouTube ID
	video, err := h.videoService.GetByID(ctx, id)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	// Try to get transcript first
	transcript, err := h.transcriptService.GetByVideoID(ctx, id)
	if err == nil && transcript != nil && transcript.Content != "" {
		// Transcript exists, use it for summary
		summary, err = h.summaryService.GenerateSummary(
			ctx,
			id,
			transcript.Content,
			"short", // default summary type
			language,
		)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, summary)
		return
	}

	// No transcript available, try audio analysis
	// Download audio and generate summary from audio
	audioPath, err := h.transcriptService.DownloadAudio(video.YouTubeID)
	if err != nil {
		// If audio download fails, try to create transcript anyway
		transcript, err := h.transcriptService.GetOrCreateTranscript(ctx, id)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
		summary, err = h.summaryService.GenerateSummary(
			ctx,
			id,
			transcript.Content,
			"short",
			language,
		)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, summary)
		return
	}

	// Clean up audio file after use
	defer func() {
		if err := os.Remove(audioPath); err != nil {
			h.logger.Warn("Failed to cleanup audio file", zap.String("path", audioPath), zap.Error(err))
		}
	}()

	// Generate summary from audio using audio analysis provider
	summary, err = h.summaryService.GenerateSummaryFromAudio(
		ctx,
		id,
		audioPath,
		"short",
		language,
	)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *VideoHandler) SummarizeVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid video ID format",
		))
		return
	}

	var req struct {
		Type      string `json:"type"`       // short, detailed, bullet_points
		FromAudio bool   `json:"from_audio"` // if true, generate summary from audio instead of transcript
		Language  string `json:"language"`   // language code for summary (e.g., "en", "tr", "auto")
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Type = "short" // default
		req.FromAudio = false
		req.Language = "" // default to empty, will use settings
	}
	if req.Language == "" {
		req.Language = "auto" // default to auto if not provided
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	// Get video to access YouTube ID
	video, err := h.videoService.GetByID(ctx, id)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	var summary *models.Summary

	if req.FromAudio {
		// Generate summary from audio using settings' whisper provider
		audioPath, err := h.transcriptService.DownloadAudio(video.YouTubeID)
		if err != nil {
			errors.HandleError(c, errors.Wrap(err, errors.ErrorCodeYouTubeDownload, errors.SubCodeYouTubeDownloadFailed, "Failed to download audio"))
			return
		}

		// Clean up audio file after use
		defer func() {
			if err := os.Remove(audioPath); err != nil {
				h.logger.Warn("Failed to cleanup audio file", zap.String("path", audioPath), zap.Error(err))
			}
		}()

		// Generate summary from audio using audio analysis provider
		summary, err = h.summaryService.GenerateSummaryFromAudio(
			ctx,
			id,
			audioPath,
			req.Type,
			req.Language,
		)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
	} else {
		// Generate summary from transcript
		// Get or create transcript first (will create if doesn't exist)
		transcript, err := h.transcriptService.GetOrCreateTranscript(ctx, id)
		if err != nil {
			errors.HandleError(c, err)
			return
		}

		// Generate summary
		summary, err = h.summaryService.GenerateSummary(
			ctx,
			id,
			transcript.Content,
			req.Type,
			req.Language,
		)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
	}

	// Update video status
	h.videoService.UpdateStatus(ctx, id, "completed")

	c.JSON(http.StatusOK, summary)
}

func (h *VideoHandler) GetSimilarVideos(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.AbortWithError(c, errors.New(
			errors.ErrorCodeBadRequest,
			errors.SubCodeInvalidInput,
			"Invalid video ID format",
		))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	minScore, _ := strconv.ParseFloat(c.DefaultQuery("min_score", "0.5"), 64)

	if limit > 50 {
		limit = 50
	}

	// Get similar videos directly from YouTube (similarity service handles this)
	similar, err := h.similarityService.FindSimilarVideos(c.Request.Context(), id, limit, minScore)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	// Get video details for each similar video
	var results []gin.H
	for _, sim := range similar {
		result := gin.H{
			"similarity_score": sim.SimilarityScore,
			"comparison_type":  sim.ComparisonType,
		}
		
		// Get video details if Video field is populated
		if sim.Video != nil {
			result["video"] = sim.Video
		} else {
			result["video"] = nil
		}
		
		results = append(results, result)
	}

	response := gin.H{
		"video_id":       id.String(),
		"similar_videos": results,
		"limit":          limit,
		"min_score":      minScore,
	}

	// YouTube always returns results, so this should not be empty
	// But if it is, return empty array without message
	if len(results) == 0 {
		response["similar_videos"] = []interface{}{} // Ensure it's an empty array, not null
	}

	c.JSON(http.StatusOK, response)
}
