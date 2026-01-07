package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // Add pprof endpoints
	"os"
	"os/signal"
	"syscall"
	"time"

	"youtube-video-summarizer/backend/internal/config"
	"youtube-video-summarizer/backend/internal/handlers"
	"youtube-video-summarizer/backend/internal/middleware"
	"youtube-video-summarizer/backend/internal/repository"
	"youtube-video-summarizer/backend/internal/services/cost"
	"youtube-video-summarizer/backend/internal/services/embedding"
	kafkaservice "youtube-video-summarizer/backend/internal/services/kafka"
	"youtube-video-summarizer/backend/internal/services/provider"
	"youtube-video-summarizer/backend/internal/services/similarity"
	settingsservice "youtube-video-summarizer/backend/internal/services/settings"
	"youtube-video-summarizer/backend/internal/services/summary"
	"youtube-video-summarizer/backend/internal/services/transcript"
	"youtube-video-summarizer/backend/internal/services/video"
	"youtube-video-summarizer/backend/internal/workers"
	"youtube-video-summarizer/backend/pkg/errors"
	"youtube-video-summarizer/backend/pkg/kafka"
	"youtube-video-summarizer/backend/pkg/youtube"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	var logger *zap.Logger
	if cfg.Server.Mode == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// Set Gin mode
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := repository.NewDatabase(repository.DatabaseConfig{URL: cfg.Database.URL}, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	
	// Get underlying sql.DB for graceful shutdown
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get database instance", zap.Error(err))
	}
	defer sqlDB.Close()

	// Initialize repositories
	videoRepo := repository.NewVideoRepository(db)
	transcriptRepo := repository.NewTranscriptRepository(db)
	summaryRepo := repository.NewSummaryRepository(db)
	embeddingRepo := repository.NewEmbeddingRepository(db)
	similarityRepo := repository.NewSimilarityRepository(db)
	costRepo := repository.NewCostRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Initialize YouTube client
	youtubeClient := youtube.NewClient(cfg.YouTube.APIKey)

	// Initialize services
	costService := cost.NewService(costRepo, logger)
	settingsService := settingsservice.NewService(settingsRepo, logger)
	
	// Initialize provider factory (manages providers based on settings)
	providerFactory := provider.NewProviderFactory(settingsService, cfg, logger)
	
	videoService := video.NewService(videoRepo, youtubeClient, logger)
	summaryService := summary.NewService(summaryRepo, providerFactory, costService, settingsService, logger, cfg)
	
	// Initialize transcript service
	transcriptService := transcript.NewService(
		transcriptRepo,
		videoRepo,
		providerFactory,
		youtubeClient,
		costService,
		logger,
	)
	
	// Initialize embedding service
	embeddingService := embedding.NewService(embeddingRepo, providerFactory, costService, transcriptService, logger)
	
	// Initialize similarity service
	similarityService := similarity.NewService(similarityRepo, embeddingRepo, videoRepo, youtubeClient, logger)

	// Initialize Kafka producer if enabled
	var kafkaProducer *kafka.Producer
	var videoEventService *kafkaservice.VideoEventService
	if cfg.Kafka.EnableKafka && len(cfg.Kafka.Brokers) > 0 {
		kafkaProducer = kafka.NewProducer(kafka.ProducerConfig{
			Brokers: cfg.Kafka.Brokers,
			Logger:  logger,
		})
		videoEventService = kafkaservice.NewVideoEventService(kafkaProducer, logger)
		logger.Info("Kafka producer initialized", zap.Strings("brokers", cfg.Kafka.Brokers))
		defer kafkaProducer.Close()
	} else {
		logger.Info("Kafka is disabled, using direct processing")
	}

	// Start Kafka workers if enabled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.Kafka.EnableKafka && len(cfg.Kafka.Brokers) > 0 && videoEventService != nil {
		// Start transcript worker
		go func() {
		if err := workers.StartTranscriptWorker(
			ctx,
			cfg.Kafka.Brokers,
			cfg.Kafka.ConsumerGroup+"-transcript",
			transcriptRepo,
			videoRepo,
			providerFactory,
			youtubeClient,
			costService,
			videoEventService,
			logger,
		); err != nil {
				logger.Error("Transcript worker failed", zap.Error(err))
			}
		}()

		// Start embedding worker
		go func() {
		if err := workers.StartEmbeddingWorker(
			ctx,
			cfg.Kafka.Brokers,
			cfg.Kafka.ConsumerGroup+"-embedding",
			embeddingRepo,
			videoRepo,
			providerFactory,
			costService,
			videoEventService,
			logger,
		); err != nil {
				logger.Error("Embedding worker failed", zap.Error(err))
			}
		}()

		// Start similarity worker
		go func() {
			if err := workers.StartSimilarityWorker(
				ctx,
				cfg.Kafka.Brokers,
				cfg.Kafka.ConsumerGroup+"-similarity",
				similarityRepo,
				embeddingRepo,
				videoRepo,
				videoEventService,
				logger,
			); err != nil {
				logger.Error("Similarity worker failed", zap.Error(err))
			}
		}()

		logger.Info("Kafka workers started",
			zap.String("consumer_group", cfg.Kafka.ConsumerGroup),
		)
	}

	// Initialize router
	router := gin.New()

	// Middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(errors.ErrorHandlerMiddleware(logger))

	// Health checks
	healthHandler := handlers.NewHealthHandler(db, logger)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Readiness)

	// pprof endpoints for performance profiling (enabled in all environments)
	router.Any("/debug/pprof/*path", gin.WrapH(http.DefaultServeMux))
	logger.Info("pprof endpoints enabled", zap.String("path", "/debug/pprof/"))

	// API routes
	api := router.Group("/api/v1")
	{
		handlers.RegisterVideoRoutes(
			api,
			videoService,
			transcriptService,
			summaryService,
			embeddingService,
			similarityService,
			videoEventService,
			logger,
		)
		handlers.RegisterSettingsRoutes(api, settingsService, cfg, logger)
		handlers.RegisterCostRoutes(api, costService, logger)
	}

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Starting server", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Cancel worker contexts
	cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
