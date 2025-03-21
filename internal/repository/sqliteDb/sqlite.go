package sqliteDb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type SQLiteRepository struct {
	Db *sql.DB
}

func New(dbPath string) (*SQLiteRepository, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Create an empty database file if it doesn't exist
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %v", err)
		}
		file.Close()
	}

	// Open SQLite database with WAL journaling and timeout settings
	db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL&_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Verify database connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables if they don't exist
	if err := initDB(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return &SQLiteRepository{Db: db}, nil
}

func initDB(db *sql.DB) error {
	// Create downloads table with enhanced schema
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS downloads (
			id INTEGER PRIMARY KEY,
			url TEXT NOT NULL,
			queue TEXT,
			file_name TEXT,
			status TEXT,
			progress INTEGER DEFAULT 0,
			headers TEXT,
			content_length INTEGER,
			content_type TEXT,
			accept_ranges BOOLEAN,
			ranges_count INTEGER,
			ranges TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create queues table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS queues (
			name TEXT PRIMARY KEY,
			folder TEXT,
			max_dl INTEGER,
			speed TEXT,
			time_range TEXT
		)
	`)
	return err
}

func (r *SQLiteRepository) Close() error {
	return r.Db.Close()
}

func (r *SQLiteRepository) SaveDownload(download models.Download) error {
	headersJSON, err := json.Marshal(download.Headers)
	if err != nil {
		log.Errorf("Error marshaling headers: %v", err)
		return err
	}

	rangesJSON, err := json.Marshal(download.Ranges)
	if err != nil {
		log.Errorf("Error marshaling ranges: %v", err)
		return err
	}

	_, err = r.Db.Exec(
		"INSERT OR REPLACE INTO downloads (id, url, queue, file_name, status, progress, headers, content_length, content_type, accept_ranges, ranges_count, ranges) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		download.Id,
		download.URL,
		download.QueueID,
		download.FileName,
		download.Status,
		download.Progress,
		string(headersJSON),
		download.ContentLength,
		download.ContentType,
		download.AcceptRanges,
		download.RangesCount,
		string(rangesJSON),
	)
	if err != nil {
		log.Errorf("Error saving download: %v", err)
		return err
	}
	return nil
}

func (r *SQLiteRepository) GetDownloads() ([]models.Download, error) {
	rows, err := r.Db.Query("SELECT id, url, queue_id, file_name, status, progress, headers, content_length, content_type, accept_ranges, ranges_count, ranges FROM downloads")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var downloads []models.Download
	for rows.Next() {
		var download models.Download
		var headersJSON, rangesJSON string
		err := rows.Scan(
			&download.Id,
			&download.URL,
			&download.QueueID,
			&download.FileName,
			&download.Status,
			&download.Progress,
			&headersJSON,
			&download.ContentLength,
			&download.ContentType,
			&download.AcceptRanges,
			&download.RangesCount,
			&rangesJSON,
		)
		if err != nil {
			return nil, err
		}

		if headersJSON != "" {
			if err := json.Unmarshal([]byte(headersJSON), &download.Headers); err != nil {
				return nil, err
			}
		}

		if rangesJSON != "" {
			if err := json.Unmarshal([]byte(rangesJSON), &download.Ranges); err != nil {
				return nil, err
			}
		}

		downloads = append(downloads, download)
	}
	return downloads, nil
}

func (r *SQLiteRepository) LoadAppState() (models.AppState, error) {
	var state models.AppState

	// Load downloads
	rows, err := r.Db.Query("SELECT url, queue, file_name, status, progress, headers FROM downloads")
	if err != nil {
		return state, err
	}
	defer rows.Close()

	for rows.Next() {
		var download models.Download
		var headersJSON string
		err := rows.Scan(&download.URL, &download.QueueID, &download.FileName, &download.Status, &download.Progress, &headersJSON)
		if err != nil {
			return state, err
		}
		// Parse headers JSON
		if headersJSON != "" {
			if err := json.Unmarshal([]byte(headersJSON), &download.Headers); err != nil {
				return state, err
			}
		}
		state.Downloads = append(state.Downloads, download)
	}

	// Load queues
	rows, err = r.Db.Query("SELECT name, folder, max_dl, speed, time_range FROM queues")
	if err != nil {
		return state, err
	}
	defer rows.Close()

	for rows.Next() {
		var queue models.Queue
		err := rows.Scan(&queue.Id, queue.Name, &queue.StorageFolder, &queue.MaxSimultaneous, &queue.BandwidthLimit, &queue.ActiveTimeStart, &queue.ActiveTimeEnd, &queue.MaxRetryAttempts)
		if err != nil {
			return state, err
		}
		state.Queues = append(state.Queues, queue)
	}

	return state, nil
}

func (r *SQLiteRepository) SaveAppState(state models.AppState) error {
	tx, err := r.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing data
	if _, err := tx.Exec("DELETE FROM downloads"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM queues"); err != nil {
		return err
	}

	// Insert downloads
	for _, download := range state.Downloads {
		headersJSON, err := json.Marshal(download.Headers)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			"INSERT INTO downloads (url, queue, file_name, status, progress, headers) VALUES (?, ?, ?, ?, ?, ?)",
			download.URL, download.QueueID, download.FileName, download.Status, download.Progress, string(headersJSON),
		)
		if err != nil {
			return err
		}
	}

	// Insert queues
	for _, queue := range state.Queues {
		_, err := tx.Exec(
			"INSERT INTO queues (name, folder, max_dl, speed, time_range) VALUES (?, ?, ?, ?, ?)",
			queue.Id, queue.Name, queue.StorageFolder, queue.MaxSimultaneous, queue.BandwidthLimit, queue.ActiveTimeStart, queue.ActiveTimeEnd, queue.MaxRetryAttempts,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
