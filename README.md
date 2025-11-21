# Shortia 🎬

Plataforma completa de generación automática de clips cortos desde videos de YouTube usando IA. Diseñada para creadores de contenido que quieren maximizar su alcance en plataformas como TikTok, YouTube Shorts e Instagram Reels.

## ✨ Características Principales

- 🎥 **Descarga inteligente** - Videos de YouTube en alta calidad con yt-dlp
- 🎤 **Transcripción automática** - OpenAI Whisper API con timestamps precisos
- 🤖 **Análisis con IA** - DeepSeek identifica los mejores momentos virales (5-8 clips sugeridos)
- 🎨 **Editor visual premium** - Interfaz moderna con glassmorphism y animaciones fluidas
- 📝 **Subtítulos personalizables** - Editor completo con plantillas, estilos y posicionamiento
- 🎬 **Renderizado en cliente** - FFmpeg.wasm procesa video directamente en el navegador
- � **SEO profesional para YouTube** - Genera títulos, descripciones y tags optimizados con IA
- ⚡ **Cache inteligente** - Redis evita regenerar contenido con DeepSeek
- 📊 **Progreso en tiempo real** - WebSockets muestran el estado de cada fase
- 💎 **Diseño moderno** - Grid animado, efectos de cristal, fuente Inter, selección púrpura

## 🛠️ Stack Tecnológico

### Frontend

- **Framework**: SvelteKit (Svelte 5) + TypeScript
- **Estilos**: Tailwind CSS con configuración personalizada
- **Tipografía**: Inter font family
- **Video Processing**: FFmpeg.wasm para renderizado en cliente
- **UI/UX**: Glassmorphism, grid animado, efectos de profundidad
- **Editor**: Canvas-based subtitle renderer con preview en tiempo real

### Backend

- **Runtime**: Go 1.21+ (Golang)
- **Framework**: Gin Web Framework
- **Base de datos**: SQLite con migraciones automáticas
- **Cache**: Redis 7-alpine con persistencia
- **WebSockets**: Gorilla WebSocket para actualizaciones en tiempo real
- **APIs**:
  - DeepSeek API para análisis de contenido y SEO
  - OpenAI Whisper API para transcripción
  - Soporte opcional para Ollama (AI local)

### Procesamiento e IA

- **FFmpeg** - Procesamiento de video y audio
- **yt-dlp** - Descarga optimizada de YouTube
- **OpenAI Whisper API** - Transcripción con timestamps
- **DeepSeek Chat** - Análisis de contenido viral y generación de SEO
- **Redis** - Cache de respuestas de IA (TTL: 24h)

### DevOps

- **Docker & Docker Compose** - Containerización completa
- **Multi-stage builds** - Imágenes optimizadas
- **Health checks** - Monitoreo de servicios
- **Persistent volumes** - Storage y cache persistentes

## 📋 Requisitos Previos

### Desarrollo Local

- **Go** 1.21 o superior
- **Node.js** 18+ y npm
- **FFmpeg** instalado en el sistema
- **yt-dlp** instalado globalmente
- **(Opcional)** Redis para cache local

### Con Docker (Recomendado)

- **Docker** 20.10+
- **Docker Compose** 2.0+

### APIs Requeridas

- **OpenAI API Key** - Para transcripción con Whisper
- **DeepSeek API Key** - Para análisis de contenido y SEO
- _Alternativa_: Ollama local (descomenta servicio en docker-compose.yml)

## 🚀 Instalación y Configuración

### Opción 1: Docker Compose (Recomendado)

1. **Clonar el repositorio**

```bash
git clone https://github.com/backsoul/shortia.git
cd shortia
```

2. **Configurar variables de entorno**

```bash
# Crea el archivo .env en la raíz
cp .env.example .env
```

Edita `.env` con tus credenciales:

```env
# === APIs de IA ===
DEEPSEEK_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxx

# === Configuración de Redis ===
REDIS_PASSWORD=
CACHE_TTL=86400

# === Opciones de IA Local (Opcional) ===
USE_OLLAMA=false
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=deepseek-r1:latest
```

3. **Levantar servicios**

```bash
docker-compose up -d --build
```

4. **Verificar servicios**

```bash
docker-compose ps
# Deberías ver: backend (8080), redis (6379)
```

5. **Acceder a la aplicación**

- Frontend: Configurar en tu servidor web o usar desarrollo local
- Backend API: http://localhost:8080
- Redis: localhost:6379

### Opción 2: Desarrollo Local

#### Backend (Go)

```bash
cd backend
go mod download
go run main.go
```

#### Frontend (SvelteKit)

```bash
cd frontend
npm install
npm run dev
```

El frontend estará en: http://localhost:5173
El backend API en: http://localhost:8080

## 🎯 Uso de la Plataforma

### 1️⃣ Subir/Procesar Video

- Pega una URL de YouTube en el input principal
- El sistema descarga automáticamente el video
- Progreso visible en tiempo real vía WebSocket

### 2️⃣ Transcripción Automática

- OpenAI Whisper transcribe el audio completo
- Genera timestamps precisos para cada segmento
- Almacena transcripción en base de datos

### 3️⃣ Análisis con IA

- DeepSeek analiza la transcripción completa
- Identifica 5-8 momentos con mayor potencial viral
- Cada clip sugerido incluye:
  - Timestamps de inicio y fin (15-60 segundos)
  - Título atractivo
  - Descripción del contenido
  - Score de viralidad (0-100)
  - Razón detallada de por qué funciona

### 4️⃣ Editar Clips

- Visualización de todos los clips sugeridos
- Editor visual con:
  - Preview del video en tiempo real
  - Subtítulos personalizables (fuente, color, posición)
  - Plantillas predefinidas
  - Canvas renderer para preview exacto
- Ajuste fino de timestamps

### 5️⃣ Generar SEO para YouTube

- Botón "Generar sugerencias de YouTube"
- DeepSeek genera profesionalmente:
  - **Título optimizado** (60-70 caracteres)
  - **Descripción rica** (600 caracteres con keywords)
  - **15-20 tags específicos** para el algoritmo
- Resultado cacheado en Redis (no regenera mismo clip)
- Un clic para copiar todos los tags

### 6️⃣ Exportar Clip

- Renderizado en cliente con FFmpeg.wasm
- Subtítulos quemados en el video
- Descarga automática en formato MP4
- Progreso visible durante exportación

## 📁 Estructura del Proyecto

```
shortia/
├── backend/                    # API en Go + Gin
│   ├── api/
│   │   ├── handlers.go        # HTTP handlers para todos los endpoints
│   │   └── websocket.go       # WebSocket para progreso en tiempo real
│   ├── services/
│   │   ├── video_service.go   # Gestión de videos y base de datos
│   │   ├── processing_service.go  # FFmpeg, yt-dlp, Whisper, DeepSeek
│   │   ├── clip_service.go    # CRUD de clips generados
│   │   ├── seo_service.go     # Generación de SEO con DeepSeek
│   │   └── cache_service.go   # Redis cache para APIs de IA
│   ├── models/
│   │   └── models.go          # Structs de Video, Clip, Transcript, SEO
│   ├── database/
│   │   └── database.go        # SQLite setup y migraciones
│   ├── Dockerfile             # Multi-stage build optimizado
│   ├── go.mod                 # Dependencias Go
│   └── main.go                # Entry point + routing
│
├── frontend/                   # SvelteKit App
│   ├── src/
│   │   ├── routes/
│   │   │   ├── +page.svelte              # Landing page con grid animado
│   │   │   ├── +layout.svelte            # Layout global
│   │   │   └── video/[id]/+page.svelte   # Vista de video individual
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   │   ├── VideoUpload.svelte           # Upload/URL input
│   │   │   │   ├── VideoList.svelte             # Lista de videos procesados
│   │   │   │   ├── VideoPlayer.svelte           # Player con controles
│   │   │   │   ├── ClipEditor.svelte            # Editor principal (2700+ líneas)
│   │   │   │   ├── SubtitleEditor.svelte        # Editor de texto de subs
│   │   │   │   ├── SubtitleSettings.svelte      # Configuración de estilo
│   │   │   │   ├── SubtitleCanvas.svelte        # Preview renderer
│   │   │   │   ├── VideoCanvasRenderer.svelte   # Canvas para export
│   │   │   │   ├── TemplateGallery.svelte       # Plantillas predefinidas
│   │   │   │   └── ExportProgress.svelte        # Barra de progreso
│   │   │   ├── services/
│   │   │   │   └── ffmpeg.ts             # FFmpeg.wasm wrapper
│   │   │   ├── data/
│   │   │   │   └── templates.ts          # Plantillas de subtítulos
│   │   │   └── types.ts                  # TypeScript interfaces
│   │   ├── app.css                       # Estilos globales + Inter font
│   │   └── app.html                      # HTML template
│   ├── static/
│   │   └── ffmpeg/                       # FFmpeg.wasm core files
│   ├── tailwind.config.js                # Tailwind con Inter
│   ├── svelte.config.js                  # SvelteKit config
│   ├── vite.config.ts                    # Vite bundler config
│   └── package.json
│
├── storage/                    # Datos persistentes (volumen Docker)
│   ├── videos/                # Videos descargados de YouTube
│   ├── clips/                 # Clips generados (legacy backend-rendered)
│   ├── transcripts/           # Transcripciones JSON
│   └── database.db            # SQLite database
│
├── binaries/                   # Binarios externos (opcional para local dev)
│   └── whisper/               # Whisper binaries (si usas local)
│
├── docker-compose.yml          # Orquestación multi-container
├── .env.example               # Template de variables de entorno
├── .env                       # Variables de entorno (no commitear)
└── README.md                  # Este archivo
```

## 🔑 API Reference

### Videos

#### `POST /api/videos`

Procesa una URL de YouTube: descarga, transcribe y analiza con IA.

**Body:**

```json
{
  "url": "https://www.youtube.com/watch?v=VIDEO_ID"
}
```

**Response:**

```json
{
  "id": "uuid-v4",
  "url": "https://youtube.com/...",
  "title": "Video Title",
  "duration": 1234.5,
  "status": "processing",
  "created_at": "2025-11-20T..."
}
```

**WebSocket:** Conectarse a `ws://localhost:8080/api/videos/{id}/ws` para recibir actualizaciones:

```json
{
  "type": "progress",
  "step": "download",
  "status": "processing",
  "message": "Descargando video..."
}
```

---

#### `GET /api/videos`

Lista todos los videos procesados.

**Response:**

```json
[
  {
    "id": "uuid",
    "title": "Video Title",
    "duration": 1234.5,
    "status": "completed",
    "thumbnail_url": "https://...",
    "created_at": "2025-11-20T..."
  }
]
```

---

#### `GET /api/videos/:id`

Obtiene detalles de un video específico.

**Response:**

```json
{
  "id": "uuid",
  "url": "https://youtube.com/...",
  "title": "Video Title",
  "duration": 1234.5,
  "status": "completed",
  "file_path": "/app/storage/videos/video.mp4",
  "thumbnail_url": "https://...",
  "created_at": "2025-11-20T..."
}
```

---

#### `GET /api/videos/:id/stream`

Stream del video para reproducción (HTTP range requests soportados).

**Response:** Video file stream

---

#### `GET /api/videos/:id/transcript`

Obtiene la transcripción completa con timestamps.

**Response:**

```json
{
  "video_id": "uuid",
  "full_text": "Transcripción completa...",
  "language": "es",
  "segments": [
    {
      "start": 0.0,
      "end": 5.2,
      "text": "Hola, bienvenidos..."
    }
  ]
}
```

---

#### `GET /api/videos/:id/clips`

Obtiene clips sugeridos por IA (5-8 mejores momentos).

**Response:**

```json
[
  {
    "start_time": 10.5,
    "end_time": 45.2,
    "title": "Cómo ganar dinero con IA",
    "description": "Explicación completa del método...",
    "score": 92,
    "reason": "Contenido viral porque..."
  }
]
```

---

#### `POST /api/videos/:id/extract-clip`

Extrae un clip raw (sin subtítulos) del video original.

**Body:**

```json
{
  "clip_start_time": 10.5,
  "clip_end_time": 45.2
}
```

**Response:**

```json
{
  "clip_path": "/app/storage/clips/clip_uuid.mp4",
  "duration": 34.7
}
```

---

#### `POST /api/videos/:id/generate-seo`

Genera SEO profesional para YouTube con DeepSeek (cacheado en Redis).

**Body:**

```json
{
  "video_id": "uuid",
  "clip_start_time": 10.5,
  "clip_end_time": 45.2,
  "clip_title": "Título del clip"
}
```

**Response:**

```json
{
  "title": "💰 Cómo GANAR DINERO con IA en 2025 | Método Paso a Paso",
  "description": "Descubre el método exacto que uso para generar ingresos con inteligencia artificial. En este clip te muestro: ✅ Las mejores herramientas ✅ Estrategias probadas ✅ Resultados reales...",
  "tags": [
    "ganar dinero online",
    "inteligencia artificial",
    "ia para emprendedores",
    "como ganar dinero con ia",
    "metodos 2025",
    ...
  ]
}
```

**Cache:** Si ya se generó SEO para los mismos parámetros, devuelve resultado cacheado instantáneamente.

---

### Clips (Legacy - Backend Processing)

#### `POST /api/clips`

Crea un clip con subtítulos procesados en backend.

**Body:**

```json
{
  "video_id": "uuid",
  "start_time": 10.5,
  "end_time": 45.2,
  "title": "Mi Clip",
  "subtitles": {
    "text": "Texto del subtítulo",
    "font_family": "Arial",
    "font_size": 24,
    "color": "#FFFFFF",
    "position": "bottom"
  }
}
```

---

#### `POST /api/clips/:id/export`

Exporta clip con subtítulos (backend processing).

---

#### `GET /api/clips/:id`

Obtiene información de un clip.

---

#### `GET /api/clips/:id/download`

Descarga el clip exportado.

---

#### `DELETE /api/clips/:id`

Elimina un clip.

---

### Utilidades

#### `POST /api/convert-webm-to-mp4`

Convierte video WebM a MP4 usando FFmpeg nativo.

**Body:** Multipart form con archivo WebM

**Response:**

```json
{
  "file_path": "/app/storage/clips/converted.mp4"
}
```

---

#### `GET /api/videos/:id/ws`

WebSocket para recibir actualizaciones de progreso en tiempo real.

**Mensajes:**

- `{"type": "progress", "step": "download", "status": "processing"}`
- `{"type": "progress", "step": "transcription", "status": "processing"}`
- `{"type": "progress", "step": "analysis", "status": "processing"}`
- `{"type": "complete", "video_id": "uuid"}`
- `{"type": "error", "message": "Error description"}`

## 🎨 Personalización de Subtítulos

El editor incluye un sistema completo de personalización de subtítulos:

### Estilos Disponibles

- **Fuentes**: Arial, Impact, Montserrat, Bebas Neue, Poppins, Roboto, etc.
- **Tamaño**: 12-120px
- **Colores**: Selector de color completo para texto, borde y fondo
- **Grosor de borde**: 0-10px
- **Opacidad de fondo**: 0-100%
- **Padding**: Espaciado interno personalizable

### Posicionamiento

- **Vertical**: Top, Center, Bottom
- **Horizontal**: Left, Center, Right
- **Offset personalizado**: Ajuste fino en píxeles

### Plantillas Predefinidas

- **Classic White** - Texto blanco con borde negro (estilo tradicional)
- **Bold Impact** - Texto amarillo con borde negro (alta visibilidad)
- **Modern Minimal** - Sin borde, fondo semitransparente
- **Neon Glow** - Efectos de brillo para contenido moderno
- **Elegant Serif** - Tipografía serif para contenido profesional

### Preview en Tiempo Real

- Canvas renderer que muestra exactamente cómo se verá el subtítulo
- Sincronización con el video player
- Ajustes instantáneos sin necesidad de exportar

## 🧠 Cómo Funciona el Sistema de IA

### 1. Análisis de Contenido (DeepSeek)

**Prompt Optimizado:**
El sistema envía la transcripción completa a DeepSeek con un prompt especializado que:

- **Prioriza coherencia narrativa**: Los clips deben tener inicio, desarrollo y cierre completo
- **Duración inteligente**: 15-60 segundos según la complejidad de la idea
- **Identifica valor**: Busca momentos que enseñen, entretengan o sorprendan
- **Evalúa potencial viral**: Score basado en curiosidad, trending topics, shareability

**Criterios de Selección:**

```
✅ Historias completas con setup y punchline
✅ Tips o consejos con explicación completa
✅ Revelaciones o insights completos
✅ Momentos emotivos con contexto
✅ Contenido trending o contra-intuitivo
```

**Output:**
Array de 5-8 clips ordenados por score de viralidad (0-100)

### 2. Generación de SEO (DeepSeek)

**Para cada clip, DeepSeek genera:**

- **Título optimizado** (60-70 caracteres):
  - Keywords en los primeros 40 caracteres
  - Emojis estratégicos para CTR
  - Gancho emocional o curiosidad
- **Descripción rica** (600 caracteres):
  - Resumen del contenido con keywords
  - Call-to-action
  - Contexto adicional
  - Hashtags relevantes
- **15-20 Tags específicos**:
  - Mix de keywords genéricas y específicas
  - Términos de búsqueda de cola larga
  - Tags trending relacionados

**Cache en Redis:**

- Key: SHA256 hash de `video_id:start_time:end_time:title`
- TTL: 24 horas (86400 segundos)
- Evita regenerar SEO para el mismo clip

### 3. Transcripción (OpenAI Whisper)

**Proceso:**

1. FFmpeg extrae audio del video descargado
2. Audio se envía a Whisper API con `response_format: verbose_json`
3. Whisper devuelve:
   - Texto completo
   - Idioma detectado
   - Segmentos con timestamps precisos (start/end)
4. Se almacena en base de datos y archivo JSON

**Ventajas de Whisper API:**

- Alta precisión en múltiples idiomas
- Timestamps precisos para sincronización
- Manejo de ruido y múltiples hablantes
- Detección automática de idioma

## ⚡ Sistema de Cache

### Redis Cache Service

**Implementación:**

```go
// cache_service.go
type CacheService struct {
    client *redis.Client
    ctx    context.Context
    ttl    time.Duration
}
```

**Métodos disponibles:**

- `GetSEO()` / `SetSEO()` - Cache de contenido SEO
- `GetTranscript()` / `SetTranscript()` - Cache de transcripciones
- `GetSuggestedClips()` / `SetSuggestedClips()` - Cache de análisis de clips
- `InvalidateVideo()` - Limpia todo el cache de un video

**Generación de Keys:**

```go
// Genera SHA256 hash para keys consistentes
func generateCacheKey(prefix string, params ...string) string {
    combined := prefix + strings.Join(params, ":")
    hash := sha256.Sum256([]byte(combined))
    return prefix + ":" + hex.EncodeToString(hash[:])
}
```

**Beneficios:**

- ✅ Reduce costos de API (DeepSeek es pagado)
- ✅ Respuestas instantáneas en cache HIT
- ✅ Persistencia entre reinicios (AOF enabled)
- ✅ TTL configurable por tipo de contenido

## 🐳 Docker Architecture

### Multi-Container Setup

```yaml
services:
  backend: # Go API + FFmpeg + yt-dlp
  redis: # Cache layer
  # ollama:  # Optional local AI
```

### Backend Container

- **Base**: `golang:1.21-alpine` (build) → `alpine:latest` (runtime)
- **Multi-stage build**: Reduce imagen final de ~1GB a ~200MB
- **Instalaciones**:
  - FFmpeg (procesamiento de video)
  - Python3 + yt-dlp (descargas de YouTube)
  - Fuentes del sistema para subtítulos
- **Volúmenes**:
  - `./storage:/app/storage` - Videos, clips, DB
  - `./binaries:/app/binaries` - Binarios opcionales

### Redis Container

- **Image**: `redis:7-alpine`
- **Persistencia**: Append-only file (AOF)
- **Volume**: `redis-data:/data`
- **Health check**: `redis-cli ping` cada 5s
- **Networking**: Accesible como `redis:6379` dentro de la red Docker

### Environment Variables

**Backend (.env):**

```env
# IA APIs
DEEPSEEK_API_KEY=sk-xxx
OPENAI_API_KEY=sk-xxx

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
CACHE_TTL=86400

# Storage
STORAGE_PATH=/app/storage
DB_PATH=/app/storage/database.db

# Processing
MAX_VIDEO_DURATION=3600
MAX_CONCURRENT_JOBS=2
WHISPER_MODEL=base
```

## 🔧 Troubleshooting

### Backend no se conecta a Redis

```bash
# Verificar que Redis está corriendo
docker-compose ps

# Ver logs de Redis
docker-compose logs redis

# Verificar conectividad
docker-compose exec backend ping redis
```

### Error de transcripción "Whisper API failed"

- Verificar `OPENAI_API_KEY` en `.env`
- Comprobar límites de API de OpenAI
- Revisar logs: `docker-compose logs backend | grep Whisper`

### DeepSeek devuelve respuestas vacías

- Aumentar `CACHE_TTL` si expira muy rápido
- Verificar `DEEPSEEK_API_KEY` válido
- Revisar logs con: `docker-compose logs backend | grep DeepSeek`
- El prompt puede ser muy largo (>16k tokens) - reducir duración de video

### FFmpeg.wasm falla en export

- Verificar que el navegador soporta SharedArrayBuffer
- Usar Chrome/Edge/Firefox moderno
- Limpiar cache del navegador
- Verificar que `static/ffmpeg/` tiene los archivos core

### Video no se descarga de YouTube

- Verificar que yt-dlp está actualizado en el container
- Algunas URLs de YouTube requieren autenticación
- Revisar logs: `docker-compose logs backend | grep yt-dlp`

## 🚀 Roadmap

### En Desarrollo

- [ ] Soporte para múltiples videos en batch
- [ ] Templates de subtítulos animados (kinetic typography)
- [ ] Export directo a TikTok/Instagram/YouTube
- [ ] Dashboard de analytics de clips exportados
- [ ] Multi-language support (UI en inglés, español, etc.)

### Futuras Mejoras

- [ ] Background music library integration
- [ ] Auto B-roll suggestions
- [ ] Face detection para auto-framing
- [ ] Voice cloning para doblajes
- [ ] A/B testing de títulos y thumbnails

## � Licencia

MIT License - Ver archivo [LICENSE](LICENSE) para más detalles.

## 🤝 Contribuciones

Las contribuciones son bienvenidas! Por favor:

1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add: AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

### Guidelines

- Mantén el código limpio y bien documentado
- Añade tests para nuevas funcionalidades
- Actualiza el README si añades endpoints o features
- Sigue los estilos de código existentes (gofmt para Go, prettier para TS)

## 👤 Autor

**backsoul**

- GitHub: [@backsoul](https://github.com/backsoul)

## ⭐ Agradecimientos

- OpenAI por Whisper API
- DeepSeek por su excelente modelo de chat
- La comunidad de SvelteKit y Go
- FFmpeg y yt-dlp por hacer posible el procesamiento de video

---

**Hecho con ❤️ para creadores de contenido que quieren maximizar su alcance en plataformas de video corto.**
