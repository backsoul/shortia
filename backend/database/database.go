package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "../storage/database.db"
	}

	// Ensure storage directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	log.Println("✅ Database initialized successfully")
	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY,
		url TEXT NOT NULL,
		title TEXT,
		duration INTEGER,
		file_path TEXT,
		thumbnail_url TEXT,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS transcripts (
		id TEXT PRIMARY KEY,
		video_id TEXT NOT NULL,
		language TEXT,
		segments TEXT, -- JSON array
		full_text TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS suggested_clips (
		id TEXT PRIMARY KEY,
		video_id TEXT NOT NULL,
		start_time REAL NOT NULL,
		end_time REAL NOT NULL,
		title TEXT,
		description TEXT,
		score REAL,
		reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS clips (
		id TEXT PRIMARY KEY,
		video_id TEXT NOT NULL,
		title TEXT,
		start_time REAL NOT NULL,
		end_time REAL NOT NULL,
		file_path TEXT,
		status TEXT DEFAULT 'processing',
		subtitles TEXT, -- JSON array
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS processing_jobs (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		progress INTEGER DEFAULT 0,
		message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
	CREATE INDEX IF NOT EXISTS idx_clips_video_id ON clips(video_id);
	CREATE INDEX IF NOT EXISTS idx_transcripts_video_id ON transcripts(video_id);
	`

	_, err := db.Exec(schema)
	return err
}
