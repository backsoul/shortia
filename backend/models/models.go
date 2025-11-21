package models

import "time"

type Video struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Duration     int       `json:"duration"`
	FilePath     string    `json:"file_path"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Status       string    `json:"status"` // pending, processing, completed, error
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Transcript struct {
	ID        string    `json:"id"`
	VideoID   string    `json:"video_id"`
	Language  string    `json:"language"`
	Segments  []Segment `json:"segments"`
	FullText  string    `json:"full_text"`
	CreatedAt time.Time `json:"created_at"`
}

type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type SuggestedClip struct {
	ID          string  `json:"id"`
	VideoID     string  `json:"video_id"`
	StartTime   float64 `json:"start_time"`
	EndTime     float64 `json:"end_time"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Reason      string  `json:"reason"`
}

type Clip struct {
	ID          string           `json:"id"`
	VideoID     string           `json:"video_id"`
	Title       string           `json:"title"`
	StartTime   float64          `json:"start_time"`
	EndTime     float64          `json:"end_time"`
	FilePath    string           `json:"file_path"`
	Status      string           `json:"status"` // processing, completed, error
	Subtitles   []SubtitleConfig `json:"subtitles"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

type SubtitleConfig struct {
	Text            string  `json:"text"`
	StartTime       float64 `json:"start_time"`
	EndTime         float64 `json:"end_time"`
	FontFamily      string  `json:"font_family"`
	FontSize        int     `json:"font_size"`
	FontWeight      int     `json:"font_weight"`
	Color           string  `json:"color"`
	BgColor         string  `json:"bg_color"`
	BgOpacity       float64 `json:"bg_opacity"`
	Position        string  `json:"position"` // top, center, bottom
	Bold            bool    `json:"bold"`
	Italic          bool    `json:"italic"`
	BorderRadius    int     `json:"border_radius"`
	ShadowBlur      int     `json:"shadow_blur"`
	Transition      string  `json:"transition"`
	ActiveTextColor string  `json:"active_text_color"`
}

type ProcessingJob struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // download, transcribe, analyze, create_clip
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
