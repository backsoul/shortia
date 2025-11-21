package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"shortgenerator/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ProcessingService struct {
	ffmpegPath  string
	ytdlpPath   string
	whisperPath string
	storagePath string
}

type fontVariant struct {
	weight int
	path   string
}

var fontLibrary = map[string][]fontVariant{
	"default": {
		{700, "/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf"},
		{600, "/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf"},
		{500, "/usr/share/fonts/dejavu/DejaVuSans.ttf"},
		{400, "/usr/share/fonts/dejavu/DejaVuSans.ttf"},
		{300, "/usr/share/fonts/liberation/LiberationSans-Regular.ttf"},
	},
	"inter": {
		{700, "/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf"},
		{600, "/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf"},
		{500, "/usr/share/fonts/dejavu/DejaVuSans.ttf"},
		{400, "/usr/share/fonts/dejavu/DejaVuSans.ttf"},
	},
	"poppins": {
		{700, "/usr/share/fonts/liberation/LiberationSans-Bold.ttf"},
		{600, "/usr/share/fonts/liberation/LiberationSans-Bold.ttf"},
		{500, "/usr/share/fonts/liberation/LiberationSans-Regular.ttf"},
		{400, "/usr/share/fonts/liberation/LiberationSans-Regular.ttf"},
	},
	"space grotesk": {
		{700, "/usr/share/fonts/liberation/LiberationSans-Bold.ttf"},
		{600, "/usr/share/fonts/liberation/LiberationSans-Bold.ttf"},
		{500, "/usr/share/fonts/liberation/LiberationSans-Regular.ttf"},
		{400, "/usr/share/fonts/liberation/LiberationSans-Regular.ttf"},
	},
	"playfair display": {
		{700, "/usr/share/fonts/freefont/FreeSerifBold.ttf"},
		{600, "/usr/share/fonts/freefont/FreeSerifBold.ttf"},
		{500, "/usr/share/fonts/freefont/FreeSerif.ttf"},
		{400, "/usr/share/fonts/freefont/FreeSerif.ttf"},
	},
	"open sans": {
		{700, "/usr/share/fonts/liberation/LiberationSans-Bold.ttf"},
		{600, "/usr/share/fonts/liberation/LiberationSans-Bold.ttf"},
		{500, "/usr/share/fonts/liberation/LiberationSans-Regular.ttf"},
		{400, "/usr/share/fonts/liberation/LiberationSans-Regular.ttf"},
	},
	"courier new": {
		{700, "/usr/share/fonts/liberation/LiberationMono-Bold.ttf"},
		{400, "/usr/share/fonts/liberation/LiberationMono-Regular.ttf"},
	},
	"monospace": {
		{700, "/usr/share/fonts/liberation/LiberationMono-Bold.ttf"},
		{400, "/usr/share/fonts/liberation/LiberationMono-Regular.ttf"},
	},
}

func NewProcessingService() *ProcessingService {
	return &ProcessingService{
		ffmpegPath:  getEnv("FFMPEG_PATH", "ffmpeg"),
		ytdlpPath:   getEnv("YTDLP_PATH", "./binaries/yt-dlp.exe"),
		whisperPath: getEnv("WHISPER_PATH", "./binaries/whisper"),
		storagePath: getEnv("STORAGE_PATH", "../storage"),
	}
}

// DownloadVideo downloads video from YouTube using yt-dlp
func (s *ProcessingService) DownloadVideo(url string, videoID string) (*models.Video, error) {
	outputPath := filepath.Join(s.storagePath, "videos", videoID+".mp4")

	// First, get metadata without downloading
	metaCmd := exec.Command(s.ytdlpPath,
		"--no-warnings",
		"--print", "title",
		"--print", "duration",
		"--print", "thumbnail",
		"--skip-download",
		url,
	)

	metaOutput, err := metaCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %v, output: %s", err, metaOutput)
	}

	// Now download the video
	downloadCmd := exec.Command(s.ytdlpPath,
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"--merge-output-format", "mp4",
		"--no-warnings",
		"--no-progress",
		"-o", outputPath,
		url,
	)

	downloadOutput, err := downloadCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %v, output: %s", err, downloadOutput)
	}

	// Parse metadata
	lines := strings.Split(strings.TrimSpace(string(metaOutput)), "\n")

	video := &models.Video{
		ID:       videoID,
		URL:      url,
		FilePath: outputPath,
	}

	if len(lines) >= 3 {
		video.Title = lines[0]
		fmt.Sscanf(lines[1], "%d", &video.Duration)
		video.ThumbnailURL = lines[2]
	}

	return video, nil
}

// TranscribeVideo generates transcript using Whisper
func (s *ProcessingService) TranscribeVideo(videoPath string, videoID string) (*models.Transcript, error) {
	// Extract audio first
	audioPath := filepath.Join(s.storagePath, "transcripts", videoID+".wav")

	cmd := exec.Command(s.ffmpegPath,
		"-y", // Overwrite output files
		"-i", videoPath,
		"-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1",
		audioPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to extract audio: %v, output: %s", err, string(output))
	}

	log.Printf("🎤 Audio extracted: %s", audioPath)

	// Use OpenAI Whisper API
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Printf("⚠️  OPENAI_API_KEY not set, using mock transcript")
		return s.createMockTranscript(videoID), nil
	}

	transcript, err := s.transcribeWithWhisperAPI(audioPath, videoID)
	if err != nil {
		log.Printf("❌ Whisper API failed: %v, using mock transcript", err)
		return s.createMockTranscript(videoID), nil
	}

	log.Printf("✅ Transcription completed: %d segments", len(transcript.Segments))
	return transcript, nil
}

func (s *ProcessingService) createMockTranscript(videoID string) *models.Transcript {
	return &models.Transcript{
		ID:       uuid.New().String(),
		VideoID:  videoID,
		Language: "en",
		Segments: []models.Segment{
			{Start: 0, End: 5, Text: "This is a sample transcript."},
			{Start: 5, End: 10, Text: "Whisper integration needed."},
		},
		FullText:  "This is a sample transcript. Whisper integration needed.",
		CreatedAt: time.Now(),
	}
}

func (s *ProcessingService) transcribeWithWhisperAPI(audioPath string, videoID string) (*models.Transcript, error) {
	apiURL := getEnv("OPENAI_API_URL", "https://api.openai.com/v1") + "/audio/transcriptions"
	apiKey := os.Getenv("OPENAI_API_KEY")

	// Open audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %v", err)
	}
	defer file.Close()

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	// Add model
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return nil, err
	}

	// Add response format with timestamps
	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return nil, err
	}

	// Close the writer
	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Make request
	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", contentType)

	log.Printf("🎤 Calling Whisper API...")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Whisper API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	log.Printf("✅ Whisper API response received")

	// Parse response
	var whisperResp struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Segments []struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
		} `json:"segments"`
	}

	if err := json.Unmarshal(respBody, &whisperResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to our format
	segments := make([]models.Segment, len(whisperResp.Segments))
	for i, seg := range whisperResp.Segments {
		segments[i] = models.Segment{
			Start: seg.Start,
			End:   seg.End,
			Text:  strings.TrimSpace(seg.Text),
		}
	}

	transcript := &models.Transcript{
		ID:        uuid.New().String(),
		VideoID:   videoID,
		Language:  whisperResp.Language,
		Segments:  segments,
		FullText:  whisperResp.Text,
		CreatedAt: time.Now(),
	}

	return transcript, nil
}

// AnalyzeTranscript uses DeepSeek to find interesting moments
func (s *ProcessingService) AnalyzeTranscript(transcript *models.Transcript, videoID string) ([]models.SuggestedClip, error) {
	useOllama := getEnv("USE_OLLAMA", "false") == "true"

	var clips []models.SuggestedClip
	var err error

	if useOllama {
		clips, err = s.analyzeWithOllama(transcript)
	} else {
		clips, err = s.analyzeWithDeepSeekAPI(transcript)
	}

	if err != nil {
		return nil, err
	}

	// Set video ID for all clips
	for i := range clips {
		clips[i].VideoID = videoID
		clips[i].ID = uuid.New().String()
	}

	return clips, nil
}

func (s *ProcessingService) analyzeWithDeepSeekAPI(transcript *models.Transcript) ([]models.SuggestedClip, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY not set")
	}

	apiURL := getEnv("DEEPSEEK_API_URL", "https://api.deepseek.com") + "/chat/completions"

	log.Printf("🤖 Calling DeepSeek API at %s", apiURL)
	log.Printf("📝 Transcript length: %d characters", len(transcript.FullText))

	prompt := fmt.Sprintf(`Eres un editor profesional de video que crea clips virales para TikTok, YouTube Shorts e Instagram Reels.

Analiza esta transcripción completa del video e identifica los 5-8 mejores momentos para crear clips cortos que tengan SENTIDO COMPLETO.

TRANSCRIPCIÓN DEL VIDEO:
%s

CRITERIOS FUNDAMENTALES PARA CADA CLIP:

1. COHERENCIA NARRATIVA (Prioridad máxima):
   - El clip DEBE tener inicio, desarrollo y cierre completo
   - NO cortar ideas a la mitad
   - NO terminar con frases incompletas
   - La historia o concepto debe ser autocontenido y comprensible

2. DURACIÓN INTELIGENTE:
   - Mínimo: 15 segundos (solo si la idea es muy concisa y potente)
   - Ideal: 30-45 segundos (suficiente para desarrollar la idea)
   - Máximo: 60 segundos (cuando sea necesario para completar el concepto)
   - PRIORIZAR la coherencia sobre la brevedad

3. PUNTOS DE INICIO Y FIN ESTRATÉGICOS:
   - Inicio: Buscar el comienzo natural de una idea, historia o concepto
   - Fin: Terminar cuando la idea esté completamente expresada
   - Evitar inicios abruptos o finales cortados
   - Respetar pausas naturales del discurso

4. CONTENIDO DE ALTO VALOR:
   - Momentos que enseñan algo específico y útil
   - Historias completas con setup y punchline
   - Revelaciones o insights completos
   - Momentos emotivos con contexto suficiente
   - Tips o consejos con explicación completa

5. POTENCIAL VIRAL:
   - Contenido que genera curiosidad desde el primer segundo
   - Información sorprendente o contra-intuitiva
   - Historias con giro o revelación
   - Contenido que invita a compartir
   - Temas trending o de interés actual

FORMATO DE RESPUESTA (JSON):
Devuelve un array de 5-8 clips con esta estructura exacta:

[
  {
    "start_time": 10.5,
    "end_time": 45.2,
    "title": "Título atractivo que refleja el contenido completo",
    "description": "Descripción detallada de lo que cubre el clip de inicio a fin",
    "score": 85,
    "reason": "Explicación específica de por qué este clip funciona: qué lo hace interesante, por qué tiene sentido completo, y su potencial viral"
  }
]

IMPORTANTE:
- Todos los textos en ESPAÑOL
- Los timestamps deben ser precisos (decimales permitidos)
- Verifica que cada clip tenga una narrativa completa
- Si un concepto necesita 50-60 segundos para completarse, úsalos
- Mejor un clip de 60 segundos coherente que uno de 30 segundos cortado
- Ordena los clips del más viral (score más alto) al menos viral

Responde ÚNICAMENTE con el array JSON, sin texto adicional.`, transcript.FullText)

	requestBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "Eres un editor profesional de video con 10+ años de experiencia creando contenido viral. Tu especialidad es identificar momentos completos y coherentes que funcionan como clips independientes. SIEMPRE priorizas que el contenido tenga sentido completo sobre la brevedad. Respondes únicamente en ESPAÑOL con JSON válido."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  4000,
		"stream":      false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	log.Printf("🚀 Sending request to DeepSeek (prompt length: %d chars)", len(prompt))

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second} // Increased timeout for long transcripts
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("✅ DeepSeek API response status: %d", resp.StatusCode)

	// Validate response body
	if len(body) == 0 {
		return nil, fmt.Errorf("DeepSeek API returned empty response body")
	}

	log.Printf("📦 DeepSeek response body length: %d bytes", len(body))
	log.Printf("📦 Response body preview (first 200 chars): %s", string(body)[:min(200, len(body))])

	// Log the response for debugging
	if resp.StatusCode != 200 {
		log.Printf("❌ DeepSeek API error response: %s", string(body))
		return nil, fmt.Errorf("DeepSeek API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("❌ JSON Unmarshal Error: %v", err)
		log.Printf("❌ Full response body: %s", string(body))
		return nil, fmt.Errorf("failed to parse response: %v, body: %s", err, string(body))
	}

	if response.Error != nil {
		return nil, fmt.Errorf("DeepSeek API error: %s (%s)", response.Error.Message, response.Error.Type)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from DeepSeek, body: %s", string(body))
	}

	// Get the content from DeepSeek response
	content := response.Choices[0].Message.Content

	if content == "" {
		return nil, fmt.Errorf("empty content from DeepSeek, full response: %s", string(body))
	}

	log.Printf("� Raw DeepSeek content length: %d characters", len(content))
	log.Printf("🎬 Content preview (first 500 chars): %s", content[:min(500, len(content))])

	// Extract JSON from markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json") + 7
		end := strings.LastIndex(content, "```")
		if end > start {
			content = content[start:end]
			content = strings.TrimSpace(content)
		}
	} else if strings.Contains(content, "```") {
		// Handle generic code blocks
		start := strings.Index(content, "```") + 3
		end := strings.LastIndex(content, "```")
		if end > start {
			content = content[start:end]
			content = strings.TrimSpace(content)
		}
	}

	log.Printf("🔍 Cleaned content for parsing (first 300 chars): %s", content[:min(300, len(content))])

	// Parse the JSON response
	var clips []models.SuggestedClip
	if err := json.Unmarshal([]byte(content), &clips); err != nil {
		log.Printf("❌ JSON Parse Error: %v", err)
		log.Printf("❌ Content that failed to parse: %s", content)
		return nil, fmt.Errorf("failed to parse DeepSeek clips response: %v", err)
	}

	log.Printf("✅ Successfully parsed %d clips from DeepSeek", len(clips))

	if len(clips) == 0 {
		return nil, fmt.Errorf("DeepSeek returned empty clips array")
	}

	return clips, nil
}

func (s *ProcessingService) analyzeWithOllama(transcript *models.Transcript) ([]models.SuggestedClip, error) {
	ollamaURL := getEnv("OLLAMA_URL", "http://localhost:11434")
	model := getEnv("OLLAMA_MODEL", "deepseek-r1:latest")

	prompt := fmt.Sprintf(`Eres un editor profesional de video que crea clips virales para TikTok, YouTube Shorts e Instagram Reels.

Analiza esta transcripción completa del video e identifica los 5-8 mejores momentos para crear clips cortos que tengan SENTIDO COMPLETO.

TRANSCRIPCIÓN DEL VIDEO:
%s

CRITERIOS FUNDAMENTALES PARA CADA CLIP:

1. COHERENCIA NARRATIVA (Prioridad máxima):
   - El clip DEBE tener inicio, desarrollo y cierre completo
   - NO cortar ideas a la mitad
   - NO terminar con frases incompletas
   - La historia o concepto debe ser autocontenido y comprensible

2. DURACIÓN INTELIGENTE:
   - Mínimo: 15 segundos (solo si la idea es muy concisa y potente)
   - Ideal: 30-45 segundos (suficiente para desarrollar la idea)
   - Máximo: 60 segundos (cuando sea necesario para completar el concepto)
   - PRIORIZAR la coherencia sobre la brevedad

3. PUNTOS DE INICIO Y FIN ESTRATÉGICOS:
   - Inicio: Buscar el comienzo natural de una idea, historia o concepto
   - Fin: Terminar cuando la idea esté completamente expresada
   - Evitar inicios abruptos o finales cortados
   - Respetar pausas naturales del discurso

4. CONTENIDO DE ALTO VALOR:
   - Momentos que enseñan algo específico y útil
   - Historias completas con setup y punchline
   - Revelaciones o insights completos
   - Momentos emotivos con contexto suficiente
   - Tips o consejos con explicación completa

5. POTENCIAL VIRAL:
   - Contenido que genera curiosidad desde el primer segundo
   - Información sorprendente o contra-intuitiva
   - Historias con giro o revelación
   - Contenido que invita a compartir
   - Temas trending o de interés actual

FORMATO DE RESPUESTA (JSON):
Devuelve un array de 5-8 clips con esta estructura exacta:

[
  {
    "start_time": 10.5,
    "end_time": 45.2,
    "title": "Título atractivo que refleja el contenido completo",
    "description": "Descripción detallada de lo que cubre el clip de inicio a fin",
    "score": 85,
    "reason": "Explicación específica de por qué este clip funciona: qué lo hace interesante, por qué tiene sentido completo, y su potencial viral"
  }
]

IMPORTANTE:
- Todos los textos en ESPAÑOL
- Los timestamps deben ser precisos (decimales permitidos)
- Verifica que cada clip tenga una narrativa completa
- Si un concepto necesita 50-60 segundos para completarse, úsalos
- Mejor un clip de 60 segundos coherente que uno de 30 segundos cortado
- Ordena los clips del más viral (score más alto) al menos viral

Responde ÚNICAMENTE con el array JSON, sin otro texto.`, transcript.FullText)

	requestBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"system": "Eres un editor profesional de video con 10+ años de experiencia creando contenido viral. Tu especialidad es identificar momentos completos y coherentes que funcionan como clips independientes. SIEMPRE priorizas que el contenido tenga sentido completo sobre la brevedad. Respondes únicamente en ESPAÑOL con JSON válido.",
	}

	jsonData, _ := json.Marshal(requestBody)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var response struct {
		Response string `json:"response"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response body: %v", err)
	}

	content := response.Response

	if content == "" {
		return nil, fmt.Errorf("empty content from Ollama, full response: %s", string(body))
	}

	log.Printf("📄 Raw Ollama content length: %d characters", len(content))
	log.Printf("🎬 Content preview (first 500 chars): %s", content[:min(500, len(content))])

	// Extract JSON from markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json") + 7
		end := strings.LastIndex(content, "```")
		if end > start {
			content = content[start:end]
			content = strings.TrimSpace(content)
		}
	} else if strings.Contains(content, "```") {
		// Handle generic code blocks
		start := strings.Index(content, "```") + 3
		end := strings.LastIndex(content, "```")
		if end > start {
			content = content[start:end]
			content = strings.TrimSpace(content)
		}
	}

	log.Printf("🔍 Cleaned content for parsing (first 300 chars): %s", content[:min(300, len(content))])

	// Parse the JSON response
	var clips []models.SuggestedClip
	if err := json.Unmarshal([]byte(content), &clips); err != nil {
		log.Printf("❌ JSON Parse Error: %v", err)
		log.Printf("❌ Content that failed to parse: %s", content)
		return nil, fmt.Errorf("failed to parse Ollama clips response: %v", err)
	}

	log.Printf("✅ Successfully parsed %d clips from Ollama", len(clips))

	if len(clips) == 0 {
		return nil, fmt.Errorf("Ollama returned empty clips array")
	}

	return clips, nil
}

// CreateClip creates a video clip with subtitles using FFmpeg
func (s *ProcessingService) CreateClip(video *models.Video, clip *models.Clip) error {
	outputPath := filepath.Join(s.storagePath, "clips", clip.ID+".mp4")

	// Ensure we have absolute paths
	inputPath := video.FilePath
	if !filepath.IsAbs(inputPath) {
		var err error
		inputPath, err = filepath.Abs(inputPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}
	}

	// Verify input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input video file does not exist: %s", inputPath)
	}

	log.Printf("📹 Creating clip from: %s", inputPath)
	log.Printf("💾 Output path: %s", outputPath)
	log.Printf("⏱️  Time: %.2f - %.2f (duration: %.2f)", clip.StartTime, clip.EndTime, clip.EndTime-clip.StartTime)
	log.Printf("📝 Subtitles: %d", len(clip.Subtitles))

	// Build FFmpeg command with subtitles
	args := []string{
		"-y",                                       // Overwrite output files
		"-ss", fmt.Sprintf("%.2f", clip.StartTime), // Start time
		"-i", inputPath, // Input file
		"-t", fmt.Sprintf("%.2f", clip.EndTime-clip.StartTime), // Duration
	}

	// Add subtitles filter if present
	if len(clip.Subtitles) > 0 {
		subtitlesFilter := s.buildSubtitlesFilter(clip.Subtitles)
		// Scale to 1080x1920 maintaining aspect ratio, then crop center if needed
		// First scale the height to 1920, then crop width to 1080 if wider
		filterComplex := fmt.Sprintf("scale=-1:1920,crop=min(iw\\,1080):1920,%s", subtitlesFilter)
		args = append(args, "-vf", filterComplex)
	} else {
		// Even without subtitles, scale and crop to 1080x1920
		args = append(args, "-vf", "scale=-1:1920,crop=min(iw\\,1080):1920")
	}

	// Output settings - maintain quality at 1080x1920 (vertical)
	args = append(args,
		"-s", "1080x1920", // Force output resolution (vertical format)
		"-c:v", "libx264", // Video codec
		"-preset", "medium", // Better quality than fast
		"-crf", "18", // High quality (lower = better, 18 is visually lossless)
		"-pix_fmt", "yuv420p", // Pixel format for compatibility
		"-c:a", "aac", // Audio codec
		"-b:a", "192k", // Higher audio bitrate
		"-movflags", "+faststart", // Enable streaming
		outputPath,
	)

	log.Printf("🎬 FFmpeg command: %s %v", s.ffmpegPath, args)

	cmd := exec.Command(s.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ FFmpeg error: %s", string(output))
		return fmt.Errorf("failed to create clip: %v, output: %s", err, output)
	}

	log.Printf("✅ Clip created successfully: %s", outputPath)

	clip.FilePath = outputPath
	clip.Status = "completed"
	now := time.Now()
	clip.CompletedAt = &now

	return nil
}

func (s *ProcessingService) buildSubtitlesFilter(subtitles []models.SubtitleConfig) string {
	// Build subtitle filters that mimic SubtitleCanvas.svelte styling as close as FFmpeg allows
	filters := []string{}

	const (
		canvasHeight     = 720.0
		bottomMargin     = 40.0
		paddingPx        = 12.0
		textShadowOffset = 1.0
		bgShadowOffset   = 4.0
	)

	scaleFactor := 1920.0 / canvasHeight // Canvas -> video scaling factor (2.666...)

	for _, sub := range subtitles {
		if strings.TrimSpace(sub.Text) == "" {
			continue
		}

		// Escape special characters for drawtext
		text := strings.ReplaceAll(sub.Text, "'", "'\\\\\\''")
		text = strings.ReplaceAll(text, ":", "\\:")

		// Font sizing (respect minimum for readability)
		fontSize := sub.FontSize
		if fontSize <= 0 {
			fontSize = 20
		}
		scaledFontSize := int(math.Round(float64(fontSize) * scaleFactor))
		if scaledFontSize < 36 {
			scaledFontSize = 36
		}

		// Resolve font path matching requested family/weight
		fontPath := resolveFontPath(sub.FontFamily, sub.FontWeight, sub.Bold)

		// Colors
		textColor := s.parseColorToFFmpeg(sub.Color, "#FFFFFF")
		bgOpacity := sub.BgOpacity
		if bgOpacity < 0 {
			bgOpacity = 0
		}
		if bgOpacity > 1 {
			bgOpacity = 1
		}
		if sub.BgOpacity == 0 && sub.BgColor == "" {
			bgOpacity = 0.8
		}
		bgColorHex := s.parseColorWithAlpha(sub.BgColor, bgOpacity)

		// Box padding (scale 12px from canvas and add subtle adjustment for radius)
		radiusAdjustment := float64(sub.BorderRadius) * 0.3
		boxBorder := int(math.Round((paddingPx + radiusAdjustment) * scaleFactor))
		if boxBorder < int(math.Round(paddingPx*scaleFactor)) {
			boxBorder = int(math.Round(paddingPx * scaleFactor))
		}

		// Positioning: match canvas middle-aligned baseline
		var targetY float64
		switch strings.ToLower(sub.Position) {
		case "top":
			targetY = bottomMargin
		case "center":
			targetY = canvasHeight / 2
		default:
			targetY = canvasHeight - bottomMargin
		}
		targetY *= scaleFactor
		yExpr := fmt.Sprintf("(%.2f)-text_h/2", targetY)
		xExpr := "(w-text_w)/2"
		enableExpr := fmt.Sprintf("enable='between(t,%.2f,%.2f)'", sub.StartTime, sub.EndTime)

		// Optional soft background shadow (simulated with offset box)
		shadowFilters := []string{}
		if bgOpacity > 0 {
			shadowBlur := sub.ShadowBlur
			if shadowBlur <= 0 {
				shadowBlur = 12
			}
			shadowSpread := int(math.Round(float64(shadowBlur) * scaleFactor / 6.0))
			shadowBorder := boxBorder + shadowSpread
			shadowOpacity := clampFloat(0.18+float64(shadowBlur)/60.0, 0.2, 0.55)
			shadowColor := fmt.Sprintf("0x000000%02X", int(math.Round(shadowOpacity*255)))
			shadowYOffset := int(math.Round(bgShadowOffset * scaleFactor))

			shadowFilter := fmt.Sprintf(
				"drawtext=text='%s':fontfile=%s:fontsize=%d:fontcolor=0x000000@0:box=1:boxcolor=%s:boxborderw=%d:x=%s:y=%s+%d:%s",
				text,
				fontPath,
				scaledFontSize,
				shadowColor,
				shadowBorder,
				xExpr,
				yExpr,
				shadowYOffset,
				enableExpr,
			)
			shadowFilters = append(shadowFilters, shadowFilter)
		}

		filters = append(filters, shadowFilters...)

		// Base text + background
		baseFilter := fmt.Sprintf(
			"drawtext=text='%s':fontfile=%s:fontsize=%d:fontcolor=%s:box=1:boxcolor=%s:boxborderw=%d:x=%s:y=%s:%s",
			text,
			fontPath,
			scaledFontSize,
			textColor,
			bgColorHex,
			boxBorder,
			xExpr,
			yExpr,
			enableExpr,
		)

		textShadowY := int(math.Round(textShadowOffset * scaleFactor))
		if textShadowY < 1 {
			textShadowY = 1
		}
		baseFilter = fmt.Sprintf("%s:shadowx=%d:shadowy=%d:shadowcolor=%s",
			baseFilter,
			0,
			textShadowY,
			"black@0.5",
		)

		filters = append(filters, baseFilter)

		// Karaoke overlay (text only)
		if sub.ActiveTextColor != "" && sub.ActiveTextColor != sub.Color {
			activeColor := s.parseColorToFFmpeg(sub.ActiveTextColor, sub.Color)
			duration := sub.EndTime - sub.StartTime
			if duration <= 0 {
				duration = 0.1
			}
			fadeDuration := duration * 0.85
			activeFilter := fmt.Sprintf(
				"drawtext=text='%s':fontfile=%s:fontsize=%d:fontcolor=%s:box=0:x=%s:y=%s:%s:alpha='min(1\\,max(0\\,(t-%.2f)/%.2f))'",
				text,
				fontPath,
				scaledFontSize,
				activeColor,
				xExpr,
				yExpr,
				enableExpr,
				sub.StartTime,
				fadeDuration,
			)

			filters = append(filters, activeFilter)
		}
	}

	return strings.Join(filters, ",")
}

// Helper to parse color to FFmpeg hex format (0xRRGGBB)
func (s *ProcessingService) parseColorToFFmpeg(color string, defaultColor string) string {
	if color == "" {
		color = defaultColor
	}

	// If color is in hex format
	if strings.HasPrefix(color, "#") {
		color = strings.TrimPrefix(color, "#")
		return "0x" + color
	}

	// If color is in rgba format, extract RGB
	if strings.HasPrefix(color, "rgba") || strings.HasPrefix(color, "rgb") {
		var r, g, b int
		if strings.HasPrefix(color, "rgba") {
			var a float64
			fmt.Sscanf(color, "rgba(%d,%d,%d,%f)", &r, &g, &b, &a)
		} else {
			fmt.Sscanf(color, "rgb(%d,%d,%d)", &r, &g, &b)
		}
		return fmt.Sprintf("0x%02X%02X%02X", r, g, b)
	}

	// Return as-is for named colors (white, black, etc.)
	return color
}

// Helper function to parse color with alpha
func (s *ProcessingService) parseColorWithAlpha(color string, opacity float64) string {
	// If color is already in rgba format
	if strings.HasPrefix(color, "rgba") {
		// Extract RGB values and apply opacity
		// Format: rgba(r, g, b, a) -> 0xRRGGBBAA
		var r, g, b int
		var a float64
		fmt.Sscanf(color, "rgba(%d,%d,%d,%f)", &r, &g, &b, &a)
		alpha := int(opacity * 255)
		return fmt.Sprintf("0x%02X%02X%02X%02X", r, g, b, alpha)
	}

	// If color is hex format
	if strings.HasPrefix(color, "#") {
		color = strings.TrimPrefix(color, "#")
		alpha := int(opacity * 255)
		return fmt.Sprintf("0x%s%02X", color, alpha)
	}

	// Default to black with opacity
	alpha := int(opacity * 255)
	return fmt.Sprintf("0x000000%02X", alpha)
}

func resolveFontPath(fontFamily string, weight int, bold bool) string {
	family := strings.ToLower(strings.TrimSpace(fontFamily))
	if family == "" {
		family = "default"
	}

	// Normalise aliases
	switch {
	case strings.Contains(family, "space") && strings.Contains(family, "grotesk"):
		family = "space grotesk"
	case strings.Contains(family, "playfair"):
		family = "playfair display"
	case strings.Contains(family, "courier"):
		family = "courier new"
	case strings.Contains(family, "mono"):
		family = "monospace"
	case strings.Contains(family, "open sans"):
		family = "open sans"
	case strings.Contains(family, "poppins"):
		family = "poppins"
	case strings.Contains(family, "inter"):
		family = "inter"
	default:
		if _, ok := fontLibrary[family]; !ok {
			family = "default"
		}
	}

	variants := fontLibrary[family]
	if len(variants) == 0 {
		return "/usr/share/fonts/dejavu/DejaVuSans.ttf"
	}

	targetWeight := weight
	if targetWeight == 0 {
		targetWeight = 400
	}
	if bold && targetWeight < 600 {
		targetWeight = 600
	}

	bestPath := variants[0].path
	bestDiff := int(math.MaxInt32)
	for _, variant := range variants {
		diff := absInt(targetWeight - variant.weight)
		if diff < bestDiff {
			bestDiff = diff
			bestPath = variant.path
		}
	}

	return bestPath
}

func clampFloat(value, minVal, maxVal float64) float64 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ExtractClipOnly - Extrae solo el clip de video sin procesarsubtítulos
// Esto permite que el frontend se encargue del rendering de subtítulos
func (s *ProcessingService) ExtractClipOnly(inputPath string, videoID string, startTime float64, endTime float64) (string, error) {
	outputDir := filepath.Join(s.storagePath, "clips")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create clips directory: %v", err)
	}

	outputFilename := fmt.Sprintf("%s_raw_%.0f-%.0f.mp4", videoID, startTime, endTime)
	outputPath := filepath.Join(outputDir, outputFilename)

	// Verificar que el video existe
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("input video file does not exist: %s", inputPath)
	}

	log.Printf("✂️ Extracting raw clip: %s", outputPath)
	log.Printf("⏱️  Time: %.2f - %.2f (duration: %.2f)", startTime, endTime, endTime-startTime)

	// Comando FFmpeg para extraer solo el clip (sin subtítulos)
	// Escalar a formato vertical 1080x1920 (9:16)
	args := []string{
		"-y",                                  // Sobrescribir si existe
		"-ss", fmt.Sprintf("%.2f", startTime), // Tiempo de inicio
		"-i", inputPath, // Video de entrada
		"-t", fmt.Sprintf("%.2f", endTime-startTime), // Duración
		"-vf", "scale=-1:1920,crop=min(iw\\,1080):1920", // Escalar a vertical
		"-s", "1080x1920", // Resolución de salida
		"-c:v", "libx264", // Codec de video
		"-preset", "fast", // Preset rápido (frontend hará el render final)
		"-crf", "18", // Alta calidad
		"-pix_fmt", "yuv420p", // Formato de píxel
		"-c:a", "aac", // Codec de audio
		"-b:a", "192k", // Bitrate de audio
		"-movflags", "+faststart", // Optimizar para streaming
		outputPath,
	}

	log.Printf("🎬 FFmpeg command: %s %v", s.ffmpegPath, args)

	cmd := exec.Command(s.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ FFmpeg error: %s", string(output))
		return "", fmt.Errorf("failed to extract clip: %v, output: %s", err, output)
	}

	log.Printf("✅ Raw clip extracted successfully: %s", outputPath)
	return outputPath, nil
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
