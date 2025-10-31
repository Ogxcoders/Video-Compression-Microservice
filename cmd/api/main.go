package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/video-compressor/internal/compressor"
	"github.com/yourusername/video-compressor/internal/database"
	"github.com/yourusername/video-compressor/internal/handlers"
	"github.com/yourusername/video-compressor/internal/middleware"
	"github.com/yourusername/video-compressor/internal/queue"
	"github.com/yourusername/video-compressor/internal/storage"
	"github.com/yourusername/video-compressor/internal/worker"
	"github.com/yourusername/video-compressor/pkg/config"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal("Invalid configuration:", err)
	}

	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		log.Fatal("Failed to create temp directory:", err)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL database")

	redisQueue, err := queue.NewRedisQueue(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisQueue.Close()
	log.Println("Connected to Redis queue")

	videoComp := compressor.NewVideoCompressor(cfg.FFmpegPath, cfg.TempDir)
	imageComp := compressor.NewImageCompressor(cfg.ImageMagickPath, cfg.TempDir)
	wpStorage := storage.NewWordPressStorage(cfg.WordPressAPIURL, cfg.WordPressUsername, cfg.WordPressAppPassword)

	w := worker.NewWorker(cfg, db, redisQueue, videoComp, imageComp, wpStorage)
	go w.Start()
	log.Println("Worker started")

	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(middleware.CORS(cfg.AllowedDomains))

	api := router.Group("/api")
	{
		api.Use(middleware.APIKeyAuth(cfg.APIKey))
		api.Use(middleware.DomainWhitelist(cfg.AllowedDomains))
		api.Use(middleware.NewRateLimiter(cfg.RateLimitPerMinute).Middleware())

		compressHandler := handlers.NewCompressHandler(db, redisQueue, cfg)

		api.POST("/compress", compressHandler.Compress)
		api.GET("/status/:job_id", compressHandler.GetStatus)
		api.GET("/result/:job_id", compressHandler.GetResult)
		api.GET("/queue/stats", compressHandler.GetQueueStats)
		api.POST("/queue/cancel/:job_id", compressHandler.CancelJob)
	}

	healthHandler := handlers.NewHealthHandler(db, redisQueue)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	log.Printf("Starting server on port %s", cfg.Port)

	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	w.Stop()
	log.Println("Server stopped")
}
