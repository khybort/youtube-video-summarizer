package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/services/cost"
)

func RegisterCostRoutes(router *gin.RouterGroup, costService *cost.Service, logger *zap.Logger) {
	handler := &CostHandler{
		costService: costService,
		logger:      logger,
	}

	costs := router.Group("/costs")
	{
		costs.GET("/summary", handler.GetCostSummary)
		costs.GET("/usage", handler.GetUsage)
		costs.GET("/videos/:id/usage", handler.GetVideoUsage)
	}
}

type CostHandler struct {
	costService CostService
	logger      *zap.Logger
}

func (h *CostHandler) GetCostSummary(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	summary, err := h.costService.GetCostSummary(c.Request.Context(), period)
	if err != nil {
		h.logger.Error("Failed to get cost summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cost summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *CostHandler) GetUsage(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	usages, err := h.costService.GetUsageByPeriod(c.Request.Context(), period)
	if err != nil {
		h.logger.Error("Failed to get usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"usage": usages})
}

func (h *CostHandler) GetVideoUsage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	usages, err := h.costService.GetUsageByVideo(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get video usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"usage": usages})
}

