package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"shortgenerator/models"
	"shortgenerator/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ProcessVideoHandler(videoService *services.VideoService, processingService *services.ProcessingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			URL string `json:"url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create video record
		video, err := videoService.CreateVideo(request.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video"})
			return
		}

		// Start processing in background
		go func() {
			// Update status to downloading
			log.Printf("📥 [%s] Starting download phase", video.ID)
			video.Status = "downloading"
			videoService.UpdateVideo(video)
			BroadcastVideoStatus(video.ID, "downloading")

			// Small delay to ensure WebSocket message is processed
			time.Sleep(500 * time.Millisecond)

			// Download video
			log.Printf("⬇️  [%s] Downloading video from URL: %s", video.ID, request.URL)
			downloadedVideo, err := processingService.DownloadVideo(request.URL, video.ID)
			if err != nil {
				log.Printf("❌ [%s] Failed to download video: %v", video.ID, err)
				video.Status = "error"
				videoService.UpdateVideo(video)
				BroadcastVideoStatus(video.ID, "error")
				return
			}

			log.Printf("✅ [%s] Video downloaded successfully: %s", video.ID, downloadedVideo.Title)

			video.Title = downloadedVideo.Title
			video.Duration = downloadedVideo.Duration
			video.FilePath = downloadedVideo.FilePath
			video.ThumbnailURL = downloadedVideo.ThumbnailURL

			// Update to transcribing phase
			log.Printf("📝 [%s] Starting transcription phase", video.ID)
			video.Status = "transcribing"
			videoService.UpdateVideo(video)
			BroadcastVideoStatus(video.ID, "transcribing")

			// Small delay to ensure WebSocket message is processed
			time.Sleep(300 * time.Millisecond)

			// Transcribe video
			log.Printf("🎤 [%s] Transcribing audio...", video.ID)
			transcript, err := processingService.TranscribeVideo(video.FilePath, video.ID)
			if err != nil {
				log.Printf("❌ [%s] Failed to transcribe video: %v", video.ID, err)
				video.Status = "error"
				videoService.UpdateVideo(video)
				BroadcastVideoStatus(video.ID, "error")
				return
			}

			log.Printf("✅ [%s] Transcription completed: %d segments", video.ID, len(transcript.Segments))

			if err := videoService.SaveTranscript(transcript); err != nil {
				log.Printf("⚠️  [%s] Failed to save transcript: %v", video.ID, err)
			}

			// Analyze with AI
			log.Printf("🤖 [%s] Starting AI analysis phase", video.ID)
			video.Status = "analyzing"
			videoService.UpdateVideo(video)
			BroadcastVideoStatus(video.ID, "analyzing")

			// Small delay to ensure WebSocket message is processed
			time.Sleep(300 * time.Millisecond)

			log.Printf("🔍 [%s] Analyzing transcript for viral clips...", video.ID)
			suggestedClips, err := processingService.AnalyzeTranscript(transcript, video.ID)
			if err != nil {
				log.Printf("❌ [%s] Failed to analyze transcript: %v", video.ID, err)
				video.Status = "error"
				videoService.UpdateVideo(video)
				BroadcastVideoStatus(video.ID, "error")
				return
			}

			log.Printf("✅ [%s] Analysis completed: %d clips suggested", video.ID, len(suggestedClips))

			if err := videoService.SaveSuggestedClips(suggestedClips); err != nil {
				log.Printf("⚠️  [%s] Failed to save suggested clips: %v", video.ID, err)
			}

			// Mark as completed
			log.Printf("🎉 [%s] All processing completed successfully!", video.ID)
			video.Status = "completed"
			videoService.UpdateVideo(video)
			BroadcastVideoStatus(video.ID, "completed")

			log.Printf("✅ [%s] Video ready: %s (Duration: %ds, %d clips)", video.ID, video.Title, video.Duration, len(suggestedClips))
		}()

		c.JSON(http.StatusOK, video)
	}
}

func GetVideosHandler(videoService *services.VideoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videos, err := videoService.GetAllVideos()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
			return
		}

		c.JSON(http.StatusOK, videos)
	}
}

func GetVideoHandler(videoService *services.VideoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		video, err := videoService.GetVideo(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		c.JSON(http.StatusOK, video)
	}
}

func GetTranscriptHandler(videoService *services.VideoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		transcript, err := videoService.GetTranscript(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transcript not found"})
			return
		}

		c.JSON(http.StatusOK, transcript)
	}
}

func GetSuggestedClipsHandler(videoService *services.VideoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		clips, err := videoService.GetSuggestedClips(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get suggested clips"})
			return
		}

		c.JSON(http.StatusOK, clips)
	}
}

func CreateClipHandler(clipService *services.ClipService, processingService *services.ProcessingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var clip models.Clip

		if err := c.ShouldBindJSON(&clip); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		clip.ID = uuid.New().String()

		if err := clipService.CreateClip(&clip); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create clip"})
			return
		}

		// Process clip in background
		go func() {
			// Get video info (you'll need to add a method to get video)
			// For now, we'll create a placeholder
			video := &models.Video{
				ID:       clip.VideoID,
				FilePath: "", // This should be fetched from DB
			}

			if err := processingService.CreateClip(video, &clip); err != nil {
				log.Printf("Failed to process clip: %v", err)
				clip.Status = "error"
				clipService.UpdateClip(&clip)
				return
			}

			clipService.UpdateClip(&clip)
			log.Printf("✅ Clip created successfully: %s", clip.ID)
		}()

		c.JSON(http.StatusOK, clip)
	}
}

func GetClipHandler(clipService *services.ClipService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		clip, err := clipService.GetClip(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Clip not found"})
			return
		}

		c.JSON(http.StatusOK, clip)
	}
}

func ExportClipHandler(videoService *services.VideoService, clipService *services.ClipService, processingService *services.ProcessingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("id")

		var request struct {
			VideoID   string                  `json:"video_id"`
			Title     string                  `json:"title"`
			StartTime float64                 `json:"start_time"`
			EndTime   float64                 `json:"end_time"`
			Subtitles []models.SubtitleConfig `json:"subtitles"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get video
		video, err := videoService.GetVideo(videoID)
		if err != nil {
			log.Printf("❌ Video not found: %s", videoID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		// Validate video file path
		if video.FilePath == "" {
			log.Printf("❌ Video file path is empty for video: %s", videoID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video file not available"})
			return
		}

		log.Printf("📹 Exporting clip from video: %s, path: %s", videoID, video.FilePath)
		log.Printf("⏱️  Time range: %.2f - %.2f", request.StartTime, request.EndTime)
		log.Printf("📝 Subtitles count: %d", len(request.Subtitles))

		// Create clip
		clip := &models.Clip{
			ID:        "",
			VideoID:   videoID,
			Title:     request.Title,
			StartTime: request.StartTime,
			EndTime:   request.EndTime,
			Subtitles: request.Subtitles,
			Status:    "processing",
		}

		if err := clipService.CreateClip(clip); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create clip"})
			return
		}

		// Process clip with subtitles
		if err := processingService.CreateClip(video, clip); err != nil {
			log.Printf("Failed to process clip: %v", err)
			clip.Status = "error"
			clipService.UpdateClip(clip)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process clip"})
			return
		}

		// Update clip status
		if err := clipService.UpdateClip(clip); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update clip"})
			return
		}

		log.Printf("✅ Clip exported successfully: %s", clip.ID)

		// Return clip info with download URL
		c.JSON(http.StatusOK, gin.H{
			"id":           clip.ID,
			"download_url": "/api/clips/" + clip.ID + "/download",
			"status":       "completed",
		})
	}
}

func DownloadClipHandler(clipService *services.ClipService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		clip, err := clipService.GetClip(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Clip not found"})
			return
		}

		if clip.Status != "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Clip is not ready yet"})
			return
		}

		// Force download instead of opening in browser
		filename := "clip_" + clip.ID + ".mp4"
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/octet-stream")
		c.File(clip.FilePath)
	}
}

func DeleteClipHandler(clipService *services.ClipService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := clipService.DeleteClip(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete clip"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Clip deleted"})
	}
}

func StreamVideoHandler(videoService *services.VideoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		video, err := videoService.GetVideo(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		if video.FilePath == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video file not found"})
			return
		}

		c.File(video.FilePath)
	}
}

// ExtractClipOnlyHandler - Nuevo endpoint simplificado que solo extrae el clip de video
// sin procesar subtítulos. El frontend se encargará de renderizar los subtítulos.
func ExtractClipOnlyHandler(videoService *services.VideoService, processingService *services.ProcessingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("id")

		var request struct {
			StartTime float64 `json:"start_time" binding:"required"`
			EndTime   float64 `json:"end_time" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validar tiempos
		if request.StartTime < 0 || request.EndTime <= request.StartTime {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range"})
			return
		}

		// Obtener video
		video, err := videoService.GetVideo(videoID)
		if err != nil {
			log.Printf("❌ Video not found: %s", videoID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		if video.FilePath == "" {
			log.Printf("❌ Video file path is empty for video: %s", videoID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video file not available"})
			return
		}

		log.Printf("✂️ Extracting raw clip from video: %s", videoID)
		log.Printf("⏱️  Time range: %.2f - %.2f (duration: %.2f)", request.StartTime, request.EndTime, request.EndTime-request.StartTime)

		// Extraer clip sin subtítulos
		clipPath, err := processingService.ExtractClipOnly(video.FilePath, videoID, request.StartTime, request.EndTime)
		if err != nil {
			log.Printf("❌ Failed to extract clip: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract clip"})
			return
		}

		log.Printf("✅ Raw clip extracted: %s", clipPath)

		// Devolver el archivo directamente para streaming
		c.Header("Content-Type", "video/mp4")
		c.Header("Accept-Ranges", "bytes")
		c.File(clipPath)
	}
}

// ConvertWebMToMP4 convierte un video WebM a MP4 usando FFmpeg nativo
func ConvertWebMToMP4(processingService *services.ProcessingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("📦 POST /api/convert-webm-to-mp4")

		// Recibir el archivo WebM
		file, err := c.FormFile("video")
		if err != nil {
			log.Printf("❌ Failed to read WebM file: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
			return
		}

		// Validar que sea WebM
		if file.Header.Get("Content-Type") != "video/webm" && !strings.HasSuffix(file.Filename, ".webm") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only WebM files are supported"})
			return
		}

		log.Printf("📥 Received WebM file: %s (%.2f MB)", file.Filename, float64(file.Size)/1024/1024)

		// Guardar el archivo temporal
		tempWebM := filepath.Join("storage", "clips", fmt.Sprintf("temp_%d.webm", time.Now().Unix()))
		if err := c.SaveUploadedFile(file, tempWebM); err != nil {
			log.Printf("❌ Failed to save WebM file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer os.Remove(tempWebM) // Limpiar archivo temporal

		log.Printf("💾 Saved temporary WebM: %s", tempWebM)

		// Convertir a MP4 usando FFmpeg nativo
		outputMP4 := filepath.Join("storage", "clips", fmt.Sprintf("converted_%d.mp4", time.Now().Unix()))

		log.Printf("🔄 Converting WebM to MP4 with native FFmpeg...")
		startTime := time.Now()

		cmd := exec.Command("ffmpeg",
			"-i", tempWebM,
			"-c:v", "libx264",
			"-preset", "medium",
			"-crf", "23",
			"-c:a", "aac",
			"-b:a", "192k",
			"-movflags", "+faststart",
			"-y",
			outputMP4,
		)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			log.Printf("❌ FFmpeg conversion failed: %v", err)
			log.Printf("FFmpeg stderr: %s", stderr.String())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Conversion failed", "details": stderr.String()})
			return
		}

		duration := time.Since(startTime)
		log.Printf("✅ Conversion complete in %.2f seconds", duration.Seconds())

		// Verificar que el archivo se creó
		if _, err := os.Stat(outputMP4); os.IsNotExist(err) {
			log.Printf("❌ Output MP4 file not found: %s", outputMP4)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Conversion failed - output not found"})
			return
		}

		// Obtener tamaño del archivo
		fileInfo, _ := os.Stat(outputMP4)
		log.Printf("📦 Output MP4 size: %.2f MB", float64(fileInfo.Size())/1024/1024)

		// Devolver el archivo MP4
		c.Header("Content-Type", "video/mp4")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(outputMP4)))
		c.File(outputMP4)

		// Limpiar archivo después de enviarlo (en una goroutine para no bloquear)
		go func() {
			time.Sleep(5 * time.Second)
			os.Remove(outputMP4)
			log.Printf("🗑️  Cleaned up temporary MP4: %s", outputMP4)
		}()
	}
}

// GenerateYouTubeSEOHandler generates professional SEO content using DeepSeek with Redis cache
func GenerateYouTubeSEOHandler(videoService *services.VideoService, cacheService *services.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			VideoID       string  `json:"video_id" binding:"required"`
			ClipStartTime float64 `json:"clip_start_time"`
			ClipEndTime   float64 `json:"clip_end_time"`
			ClipTitle     string  `json:"clip_title"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("❌ SEO request binding error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("🎯 [%s] Generating SEO for clip: %s (%.1fs - %.1fs)",
			request.VideoID, request.ClipTitle, request.ClipStartTime, request.ClipEndTime)

		// Try to get from cache first
		cachedSEO, err := cacheService.GetSEO(request.VideoID, request.ClipStartTime, request.ClipEndTime, request.ClipTitle)
		if err == nil && cachedSEO != nil {
			log.Printf("✅ [%s] Returning cached SEO", request.VideoID)
			c.JSON(http.StatusOK, cachedSEO)
			return
		}

		// Get video
		video, err := videoService.GetVideo(request.VideoID)
		if err != nil {
			log.Printf("❌ [%s] Video not found for SEO generation", request.VideoID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		// Get transcript
		transcript, err := videoService.GetTranscript(request.VideoID)
		if err != nil {
			log.Printf("❌ [%s] Failed to get transcript for SEO generation: %v", request.VideoID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transcript"})
			return
		}

		// Extract relevant transcript segments for the clip
		var clipText strings.Builder
		segmentsFound := 0
		for _, segment := range transcript.Segments {
			if segment.Start >= request.ClipStartTime && segment.End <= request.ClipEndTime {
				clipText.WriteString(segment.Text)
				clipText.WriteString(" ")
				segmentsFound++
			}
		}

		transcriptText := strings.TrimSpace(clipText.String())
		if transcriptText == "" {
			transcriptText = "Video clip"
			log.Printf("⚠️  [%s] No transcript segments found for clip range %.1f-%.1f, using fallback",
				request.VideoID, request.ClipStartTime, request.ClipEndTime)
		} else {
			log.Printf("📝 [%s] Extracted %d transcript segments (%d characters) for SEO generation",
				request.VideoID, segmentsFound, len(transcriptText))
		}

		// Call DeepSeek API to generate professional SEO content
		deepseekAPIKey := os.Getenv("DEEPSEEK_API_KEY")
		if deepseekAPIKey == "" {
			log.Printf("❌ DEEPSEEK_API_KEY not configured")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DeepSeek API key not configured"})
			return
		}

		log.Printf("🤖 [%s] Calling DeepSeek API for SEO generation...", request.VideoID)
		seoContent, err := services.GenerateProfessionalSEO(deepseekAPIKey, transcriptText, request.ClipTitle, video.Title)
		if err != nil {
			log.Printf("❌ [%s] Failed to generate SEO: %v", request.VideoID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate SEO content"})
			return
		}

		// Cache the generated SEO content
		if err := cacheService.SetSEO(request.VideoID, request.ClipStartTime, request.ClipEndTime, request.ClipTitle, seoContent); err != nil {
			log.Printf("⚠️ [%s] Failed to cache SEO: %v", request.VideoID, err)
		}

		log.Printf("✅ [%s] SEO generated successfully: Title=%s, Tags=%d",
			request.VideoID, seoContent.Title, len(seoContent.Tags))
		c.JSON(http.StatusOK, seoContent)
	}
}
