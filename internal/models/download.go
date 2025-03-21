package models

import (
	"net/http"
	"time"
)

type DownloadStatus string

const (
	DownloadStatusDownloading DownloadStatus = "DOWNLOADING"
	DownloadStatusPaused      DownloadStatus = "PAUSED"
	DownloadStatusCompleted   DownloadStatus = "COMPLETED"
	DownloadStatusCanceled    DownloadStatus = "CANCELED"
	DownloadStatusFailed      DownloadStatus = "FAILED"
	DownloadStatusQueued      DownloadStatus = "QUEUED"
)

type Download struct {
	Id            int64          `json:"id" sqliteDb:"id,primary"`
	URL           string         `json:"url" sqliteDb:"url"`
	QueueID       int64          `json:"queue_id" sqliteDb:"queue_id"`
	QueueName     string         `json:"queue_name" sqliteDb:"queue_name"`
	FileName      string         `json:"file_name" sqliteDb:"file_name"`
	Status        DownloadStatus `json:"status" sqliteDb:"status"`
	Progress      int            `json:"progress" sqliteDb:"progress"`
	Headers       http.Header    `json:"headers" sqliteDb:"headers"`
	ContentLength int64          `json:"content_length" sqliteDb:"content_length"`
	ContentType   string         `json:"content_type" sqliteDb:"content_type"`
	AcceptRanges  bool           `json:"accept_ranges" sqliteDb:"accept_ranges"`
	RangesCount   int            `json:"ranges_count" sqliteDb:"ranges_count"`
	Ranges        []int          `json:"ranges" sqliteDb:"ranges"`
	// Exported fields for progress tracking.
	CurrentProgress int64     // Bytes downloaded so far.
	StartTime       time.Time // When the download started.
	// Internal fields.
	CancelChan chan struct{}
}

type Queue struct {
	Id               int64  `json:"id" sqliteDb:"id,primary"`
	Name             string `json:"name" sqliteDb:"name"`
	StorageFolder    string `json:"storage_folder" sqliteDb:"storage_folder"`
	MaxSimultaneous  int    `json:"max_simultaneous" sqliteDb:"max_simultaneous"`
	BandwidthLimit   int64  `json:"bandwidth_limit" sqliteDb:"bandwidth_limit"`
	MaxDownloadSpeed int64  `json:"max_download_speed" sqliteDb:"max_download_speed"`
	ActiveTimeStart  string `json:"active_time_start" sqliteDb:"active_time_start"`
	ActiveTimeEnd    string `json:"active_time_end" sqliteDb:"active_time_end"`
	MaxRetryAttempts int    `json:"max_retry_attempts" sqliteDb:"max_retry_attempts"`
}
