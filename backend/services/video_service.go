package services

import (
	"database/sql"
	"encoding/json"
	"shortgenerator/models"
	"time"

	"github.com/google/uuid"
)

type VideoService struct {
	db *sql.DB
}

func NewVideoService(db *sql.DB) *VideoService {
	return &VideoService{db: db}
}

func (s *VideoService) CreateVideo(url string) (*models.Video, error) {
	video := &models.Video{
		ID:        uuid.New().String(),
		URL:       url,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `INSERT INTO videos (id, url, status, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, video.ID, video.URL, video.Status, video.CreatedAt, video.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return video, nil
}

func (s *VideoService) GetVideo(id string) (*models.Video, error) {
	video := &models.Video{}
	query := `SELECT id, url, title, duration, file_path, thumbnail_url, status, created_at, updated_at 
			  FROM videos WHERE id = ?`
	
	err := s.db.QueryRow(query, id).Scan(
		&video.ID, &video.URL, &video.Title, &video.Duration,
		&video.FilePath, &video.ThumbnailURL, &video.Status,
		&video.CreatedAt, &video.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return video, nil
}

func (s *VideoService) GetAllVideos() ([]models.Video, error) {
	query := `SELECT id, url, title, duration, file_path, thumbnail_url, status, created_at, updated_at 
			  FROM videos ORDER BY created_at DESC`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	videos := []models.Video{}
	for rows.Next() {
		var video models.Video
		err := rows.Scan(
			&video.ID, &video.URL, &video.Title, &video.Duration,
			&video.FilePath, &video.ThumbnailURL, &video.Status,
			&video.CreatedAt, &video.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	return videos, nil
}

func (s *VideoService) UpdateVideo(video *models.Video) error {
	query := `UPDATE videos 
			  SET title = ?, duration = ?, file_path = ?, thumbnail_url = ?, status = ?, updated_at = ?
			  WHERE id = ?`
	
	_, err := s.db.Exec(query, video.Title, video.Duration, video.FilePath,
		video.ThumbnailURL, video.Status, time.Now(), video.ID)
	
	return err
}

func (s *VideoService) SaveTranscript(transcript *models.Transcript) error {
	segmentsJSON, err := json.Marshal(transcript.Segments)
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO transcripts (id, video_id, language, segments, full_text, created_at)
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err = s.db.Exec(query, transcript.ID, transcript.VideoID, transcript.Language,
		string(segmentsJSON), transcript.FullText, time.Now())
	
	return err
}

func (s *VideoService) GetTranscript(videoID string) (*models.Transcript, error) {
	transcript := &models.Transcript{}
	var segmentsJSON string

	query := `SELECT id, video_id, language, segments, full_text, created_at 
			  FROM transcripts WHERE video_id = ?`
	
	err := s.db.QueryRow(query, videoID).Scan(
		&transcript.ID, &transcript.VideoID, &transcript.Language,
		&segmentsJSON, &transcript.FullText, &transcript.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(segmentsJSON), &transcript.Segments); err != nil {
		return nil, err
	}

	return transcript, nil
}

func (s *VideoService) SaveSuggestedClips(clips []models.SuggestedClip) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO suggested_clips (id, video_id, start_time, end_time, title, description, score, reason)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	for _, clip := range clips {
		_, err := tx.Exec(query, clip.ID, clip.VideoID, clip.StartTime, clip.EndTime,
			clip.Title, clip.Description, clip.Score, clip.Reason)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *VideoService) GetSuggestedClips(videoID string) ([]models.SuggestedClip, error) {
	query := `SELECT id, video_id, start_time, end_time, title, description, score, reason
			  FROM suggested_clips WHERE video_id = ? ORDER BY score DESC`
	
	rows, err := s.db.Query(query, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clips := []models.SuggestedClip{}
	for rows.Next() {
		var clip models.SuggestedClip
		err := rows.Scan(&clip.ID, &clip.VideoID, &clip.StartTime, &clip.EndTime,
			&clip.Title, &clip.Description, &clip.Score, &clip.Reason)
		if err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}

	return clips, nil
}
