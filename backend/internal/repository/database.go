package repository

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"go.uber.org/zap"
	"youtube-video-summarizer/backend/internal/models"
)

func NewDatabase(cfg DatabaseConfig, zapLogger *zap.Logger) (*gorm.DB, error) {
	// Create GORM logger from zap logger
	gormLogger := logger.New(
		&zapLogAdapter{logger: zapLogger},
		logger.Config{
			SlowThreshold:             200,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for ping
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	zapLogger.Info("Database connection established")

	// Enable extensions
	if err := enableExtensions(db, zapLogger); err != nil {
		return nil, fmt.Errorf("failed to enable extensions: %w", err)
	}

	// Run migrations
	if err := runMigrations(db, zapLogger); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

type DatabaseConfig struct {
	URL string
}

// zapLogAdapter adapts zap logger to GORM logger interface
type zapLogAdapter struct {
	logger *zap.Logger
}

func (z *zapLogAdapter) Printf(format string, v ...interface{}) {
	z.logger.Info(fmt.Sprintf(format, v...))
}

func enableExtensions(db *gorm.DB, zapLogger *zap.Logger) error {
	extensions := []string{
		"CREATE EXTENSION IF NOT EXISTS vector",
		"CREATE EXTENSION IF NOT EXISTS pg_trgm",
	}

	for _, ext := range extensions {
		if err := db.Exec(ext).Error; err != nil {
			zapLogger.Warn("Failed to create extension", zap.String("extension", ext), zap.Error(err))
		}
	}

	return nil
}

func runMigrations(db *gorm.DB, zapLogger *zap.Logger) error {
	// AutoMigrate all models
	err := db.AutoMigrate(
		&models.Video{},
		&models.Transcript{},
		&models.Summary{},
		&models.VideoEmbedding{},
		&models.VideoSimilarity{},
		&models.TokenUsage{},
		&models.Settings{},
	)
	if err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	// Create unique constraint for video_embeddings (video_id, embedding_type)
	uniqueConstraint := `CREATE UNIQUE INDEX IF NOT EXISTS idx_embeddings_video_type_unique ON video_embeddings(video_id, embedding_type)`
	if err := db.Exec(uniqueConstraint).Error; err != nil {
		zapLogger.Warn("Failed to create unique constraint for embeddings", zap.Error(err))
	}

	// Create custom indexes that GORM might not create
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_videos_youtube_id ON videos(youtube_id)",
		"CREATE INDEX IF NOT EXISTS idx_videos_channel_id ON videos(channel_id)",
		"CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status)",
		"CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_transcripts_video_id ON transcripts(video_id)",
		"CREATE INDEX IF NOT EXISTS idx_summaries_video_id ON summaries(video_id)",
		"CREATE INDEX IF NOT EXISTS idx_embeddings_video_id ON video_embeddings(video_id)",
		"CREATE INDEX IF NOT EXISTS idx_embeddings_type ON video_embeddings(embedding_type)",
		"CREATE INDEX IF NOT EXISTS idx_similarities_video1 ON video_similarities(video_id_1)",
		"CREATE INDEX IF NOT EXISTS idx_similarities_video2 ON video_similarities(video_id_2)",
		"CREATE INDEX IF NOT EXISTS idx_similarities_combined ON video_similarities(combined_similarity DESC)",
		"CREATE INDEX IF NOT EXISTS idx_token_usage_video_id ON token_usage(video_id)",
		"CREATE INDEX IF NOT EXISTS idx_token_usage_provider ON token_usage(provider)",
		"CREATE INDEX IF NOT EXISTS idx_token_usage_operation ON token_usage(operation)",
		"CREATE INDEX IF NOT EXISTS idx_token_usage_created_at ON token_usage(created_at DESC)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			zapLogger.Warn("Failed to create index", zap.String("index", idx), zap.Error(err))
		}
	}

	// Create HNSW index for embeddings (pgvector)
	hnswIndex := `CREATE INDEX IF NOT EXISTS idx_embeddings_hnsw ON video_embeddings USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)`
	if err := db.Exec(hnswIndex).Error; err != nil {
		zapLogger.Warn("Failed to create HNSW index for embeddings", zap.Error(err))
	}

	// Create unique index for video similarities
	uniqueIndex := `CREATE UNIQUE INDEX IF NOT EXISTS idx_similarities_unique_pair ON video_similarities(LEAST(video_id_1, video_id_2), GREATEST(video_id_1, video_id_2))`
	if err := db.Exec(uniqueIndex).Error; err != nil {
		zapLogger.Warn("Failed to create unique index for similarities", zap.Error(err))
	}

	zapLogger.Info("Migrations completed successfully")
	return nil
}
