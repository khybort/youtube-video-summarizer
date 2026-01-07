package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"youtube-video-summarizer/backend/internal/models"
)

type CostRepository interface {
	Create(ctx context.Context, usage *models.TokenUsage) error
	GetByVideoID(ctx context.Context, videoID uuid.UUID) ([]*models.TokenUsage, error)
	GetSummary(ctx context.Context, startDate, endDate *time.Time) (*models.CostSummary, error)
	GetByPeriod(ctx context.Context, period string) ([]*models.TokenUsage, error)
}

type costRepository struct {
	db *gorm.DB
}

func NewCostRepository(db *gorm.DB) CostRepository {
	return &costRepository{db: db}
}

func (r *costRepository) Create(ctx context.Context, usage *models.TokenUsage) error {
	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *costRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID) ([]*models.TokenUsage, error) {
	var usages []*models.TokenUsage
	err := r.db.WithContext(ctx).
		Where("video_id = ?", videoID).
		Order("created_at DESC").
		Find(&usages).Error
	return usages, err
}

func (r *costRepository) GetSummary(ctx context.Context, startDate, endDate *time.Time) (*models.CostSummary, error) {
	query := r.db.WithContext(ctx).Model(&models.TokenUsage{})

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	var summary models.CostSummary
	summary.ByProvider = make(map[string]float64)
	summary.ByOperation = make(map[string]float64)
	summary.ByModel = make(map[string]float64)

	// Get totals
	var result struct {
		TotalCost   float64
		TotalTokens int64
		VideoCount  int64
	}
	query.Select("COALESCE(SUM(cost), 0) as total_cost, COALESCE(SUM(total_tokens), 0) as total_tokens, COUNT(DISTINCT video_id) as video_count").
		Scan(&result)

	summary.TotalCost = result.TotalCost
	summary.TotalTokens = int(result.TotalTokens)
	summary.VideoCount = int(result.VideoCount)

	// Get by provider
	var providerResults []struct {
		Provider string
		Cost     float64
	}
	query.Group("provider").Select("provider, SUM(cost) as cost").Scan(&providerResults)
	for _, p := range providerResults {
		summary.ByProvider[p.Provider] = p.Cost
	}

	// Get by operation
	var operationResults []struct {
		Operation string
		Cost      float64
	}
	query.Group("operation").Select("operation, SUM(cost) as cost").Scan(&operationResults)
	for _, o := range operationResults {
		summary.ByOperation[o.Operation] = o.Cost
	}

	// Get by model
	var modelResults []struct {
		Model string
		Cost  float64
	}
	query.Group("model").Select("model, SUM(cost) as cost").Scan(&modelResults)
	for _, m := range modelResults {
		summary.ByModel[m.Model] = m.Cost
	}

	if summary.VideoCount > 0 {
		summary.AverageCostPerVideo = summary.TotalCost / float64(summary.VideoCount)
	}

	return &summary, nil
}

func (r *costRepository) GetByPeriod(ctx context.Context, period string) ([]*models.TokenUsage, error) {
	var startDate time.Time
	now := time.Now()

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "all":
		// No date filter
		var usages []*models.TokenUsage
		err := r.db.WithContext(ctx).
			Order("created_at DESC").
			Find(&usages).Error
		return usages, err
	default:
		startDate = now.AddDate(0, -1, 0) // Default to month
	}

	var usages []*models.TokenUsage
	err := r.db.WithContext(ctx).
		Where("created_at >= ?", startDate).
		Order("created_at DESC").
		Find(&usages).Error

	return usages, err
}
