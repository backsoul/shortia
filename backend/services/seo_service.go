package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SEOContent struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type DeepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []DeepSeekMessage `json:"messages"`
	Stream   bool              `json:"stream"`
}

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// GenerateProfessionalSEO generates professional YouTube SEO content using DeepSeek
func GenerateProfessionalSEO(apiKey, transcriptText, clipTitle, videoTitle string) (*SEOContent, error) {
	prompt := fmt.Sprintf(`Eres un experto en SEO de YouTube con más de 10 años de experiencia optimizando contenido para posicionamiento orgánico.

CONTEXTO DEL VIDEO:
Título del video original: %s
Título del clip: %s
Transcripción del clip: %s

TAREA:
Genera contenido SEO profesional optimizado para YouTube Shorts/clips verticales siguiendo estas especificaciones exactas:

1. TÍTULO (máximo 100 caracteres):
   - Debe ser atractivo y contener palabras clave relevantes
   - Incluir números o datos cuando sea posible (ej: "5 formas de...")
   - Usar palabras de poder: "Cómo", "Por qué", "Mejor", "Secreto", etc.
   - NO usar emojis excesivos (máximo 1)
   - Debe generar curiosidad pero ser honesto con el contenido

2. DESCRIPCIÓN (exactamente 600 caracteres):
   - Primera línea: Hook potente que capture atención
   - Párrafo principal: Explicar el contenido con palabras clave naturalmente integradas
   - Incluir call-to-action sutil
   - Usar espaciado estratégico para mejorar legibilidad
   - Incluir hashtags relevantes AL FINAL (#shorts #viral #trending)
   - Optimizada para búsqueda semántica de YouTube

3. TAGS (15-20 tags específicos):
   - NO usar tags genéricos como "video", "content", "viral"
   - Usar long-tail keywords específicos del nicho
   - Incluir variaciones del tema principal
   - Mezclar tags de alto y medio volumen de búsqueda
   - Incluir términos técnicos si aplica
   - Tags en español e inglés cuando sea relevante

FORMATO DE RESPUESTA (JSON estricto):
{
  "title": "Título optimizado aquí",
  "description": "Descripción de exactamente 600 caracteres aquí",
  "tags": ["tag1", "tag2", "tag3", ...]
}

IMPORTANTE: Responde ÚNICAMENTE con el JSON, sin texto adicional antes o después.`, videoTitle, clipTitle, transcriptText)

	reqBody := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{
				Role:    "system",
				Content: "Eres un experto en SEO de YouTube. Respondes únicamente en formato JSON válido sin texto adicional.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call DeepSeek API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DeepSeek API error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var deepseekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to parse DeepSeek response: %v", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in DeepSeek response")
	}

	content := deepseekResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// Remove markdown code blocks if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var seoContent SEOContent
	if err := json.Unmarshal([]byte(content), &seoContent); err != nil {
		return nil, fmt.Errorf("failed to parse SEO content: %v - content: %s", err, content)
	}

	return &seoContent, nil
}
