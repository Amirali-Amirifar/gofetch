package models

import "net/http"

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
	Queue         string         `json:"queue" sqliteDb:"queue"`
	FileName      string         `json:"file_name" sqliteDb:"file_name"`
	Status        DownloadStatus `json:"status" sqliteDb:"status"`
	Progress      int            `json:"progress" sqliteDb:"progress"`
	Headers       http.Header    `json:"headers" sqliteDb:"headers"`
	ContentLength int64          `json:"content_length" sqliteDb:"content_length"`
	ContentType   string         `json:"content_type" sqliteDb:"content_type"`
	AcceptRanges  bool           `json:"accept_ranges" sqliteDb:"accept_ranges"`
	RangesCount   int            `json:"ranges_count" sqliteDb:"ranges_count"`
	Ranges        []int          `json:"ranges" sqliteDb:"ranges"`
}
