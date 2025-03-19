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
	URL      string         `json:"url"`
	Queue    string         `json:"queue"`
	FileName string         `json:"file_name"`
	Status   DownloadStatus `json:"status"`
	Progress int            `json:"progress"`
	Headers  http.Header    `json:"headers"`
}
