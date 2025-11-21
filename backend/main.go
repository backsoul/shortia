package main

import (
	"log"
	"os"

	"shortgenerator/api"
	"shortgenerator/database"
	"shortgenerator/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize services
	videoService := services.NewVideoService(db)
	clipService := services.NewClipService(db)
	processingService := services.NewProcessingService()
	cacheService := services.NewCacheService()
	defer cacheService.Close()

	// Setup Gin router
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:4173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(config))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	apiRouter := router.Group("/api")
	{
		// Videos
		apiRouter.POST("/videos", api.ProcessVideoHandler(videoService, processingService))
		apiRouter.GET("/videos", api.GetVideosHandler(videoService))
		apiRouter.GET("/videos/:id", api.GetVideoHandler(videoService))
		apiRouter.GET("/videos/:id/stream", api.StreamVideoHandler(videoService))
		apiRouter.GET("/videos/:id/transcript", api.GetTranscriptHandler(videoService))
		apiRouter.GET("/videos/:id/clips", api.GetSuggestedClipsHandler(videoService))

		// NEW: Extract raw clip without subtitles (frontend will handle rendering)
		apiRouter.POST("/videos/:id/extract-clip", api.ExtractClipOnlyHandler(videoService, processingService))

		// NEW: Convert WebM to MP4 using native FFmpeg
		apiRouter.POST("/convert-webm-to-mp4", api.ConvertWebMToMP4(processingService))

		// NEW: Generate professional YouTube SEO using DeepSeek (with Redis cache)
		apiRouter.POST("/videos/:id/generate-seo", api.GenerateYouTubeSEOHandler(videoService, cacheService))

		// Clips (legacy - con subtítulos procesados en backend)
		apiRouter.POST("/clips", api.CreateClipHandler(clipService, processingService))
		apiRouter.POST("/clips/:id/export", api.ExportClipHandler(videoService, clipService, processingService))
		apiRouter.GET("/clips/:id", api.GetClipHandler(clipService))
		apiRouter.GET("/clips/:id/download", api.DownloadClipHandler(clipService))
		apiRouter.DELETE("/clips/:id", api.DeleteClipHandler(clipService))

		// WebSocket for progress updates (video-specific)
		apiRouter.GET("/videos/:id/ws", api.VideoWebSocketHandler())
	}

	// Static files for downloaded clips
	router.Static("/downloads", "./storage/clips")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
