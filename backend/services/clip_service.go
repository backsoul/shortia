package services

import (
	"database/sql"
	"encoding/json"
	"shortgenerator/models"
	"time"

	"github.com/google/uuid"
)

type ClipService struct {
	db *sql.DB
}

func NewClipService(db *sql.DB) *ClipService {
	return &ClipService{db: db}
}

func (s *ClipService) CreateClip(clip *models.Clip) error {
	if clip.ID == "" {
		clip.ID = uuid.New().String()
	}
	clip.CreatedAt = time.Now()
	clip.Status = "processing"

	subtitlesJSON, err := json.Marshal(clip.Subtitles)
	if err != nil {
		return err
	}

	query := `INSERT INTO clips (id, video_id, title, start_time, end_time, status, subtitles, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err = s.db.Exec(query, clip.ID, clip.VideoID, clip.Title, clip.StartTime,
		clip.EndTime, clip.Status, string(subtitlesJSON), clip.CreatedAt)
	
	return err
}

func (s *ClipService) GetClip(id string) (*models.Clip, error) {
	clip := &models.Clip{}
	var subtitlesJSON string
	var completedAt sql.NullTime

	query := `SELECT id, video_id, title, start_time, end_time, file_path, status, subtitles, created_at, completed_at
			  FROM clips WHERE id = ?`
	
	err := s.db.QueryRow(query, id).Scan(
		&clip.ID, &clip.VideoID, &clip.Title, &clip.StartTime, &clip.EndTime,
		&clip.FilePath, &clip.Status, &subtitlesJSON, &clip.CreatedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		clip.CompletedAt = &completedAt.Time
	}

	if err := json.Unmarshal([]byte(subtitlesJSON), &clip.Subtitles); err != nil {
		return nil, err
	}

	return clip, nil
}

func (s *ClipService) UpdateClip(clip *models.Clip) error {
	subtitlesJSON, err := json.Marshal(clip.Subtitles)
	if err != nil {
		return err
	}

	query := `UPDATE clips 
			  SET file_path = ?, status = ?, subtitles = ?, completed_at = ?
			  WHERE id = ?`
	
	_, err = s.db.Exec(query, clip.FilePath, clip.Status, string(subtitlesJSON),
		clip.CompletedAt, clip.ID)
	
	return err
}

func (s *ClipService) DeleteClip(id string) error {
	query := `DELETE FROM clips WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *ClipService) GetClipsByVideo(videoID string) ([]models.Clip, error) {
	query := `SELECT id, video_id, title, start_time, end_time, file_path, status, subtitles, created_at, completed_at
			  FROM clips WHERE video_id = ? ORDER BY created_at DESC`
	
	rows, err := s.db.Query(query, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clips := []models.Clip{}
	for rows.Next() {
		var clip models.Clip
		var subtitlesJSON string
		var completedAt sql.NullTime

		err := rows.Scan(&clip.ID, &clip.VideoID, &clip.Title, &clip.StartTime, &clip.EndTime,
			&clip.FilePath, &clip.Status, &subtitlesJSON, &clip.CreatedAt, &completedAt)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			clip.CompletedAt = &completedAt.Time
		}

		if err := json.Unmarshal([]byte(subtitlesJSON), &clip.Subtitles); err != nil {
			return nil, err
		}

		clips = append(clips, clip)
	}

	return clips, nil
}
