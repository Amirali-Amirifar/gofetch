package controller

import (
	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// QueueManager represents a queue of downloads with constraints on simultaneous downloads
// and active time range
type QueueManager struct {
	models.Queue
	ActiveDownloads int           // Current active downloads
	DownloadChannel chan struct{} // Channel for managing active downloads
	mu              sync.Mutex    // Mutex to ensure safe concurrent access to ActiveDownloads
	DownloadLimiter chan struct{} // Token channel for limiting download speed (throttling)
}

// CanStartDownload checks whether a new download can be started
// based on the time range and max simultaneous downloads constraints
func (q *QueueManager) CanStartDownload() bool {
	currentTime := time.Now()
	start, _ := time.Parse("15:04", q.ActiveTimeStart)
	end, _ := time.Parse("15:04", q.ActiveTimeEnd)

	if q.ActiveTimeStart != "" && q.ActiveTimeEnd != "" {
		if currentTime.Before(start) || currentTime.After(end) {
			return false
		}
	}

	// Locking the mutex to safely check and modify the number of active downloads
	q.mu.Lock()
	canStart := q.ActiveDownloads < q.MaxSimultaneous
	q.mu.Unlock()

	return canStart
}

// StartDownload initiates a download process if allowed by the queue constraints
func (q *QueueManager) StartDownload(d *Download) {
	if !q.CanStartDownload() {
		log.Infof("Download %s is queued due to queue restrictions", d.URL)
		d.Status = models.DownloadStatusQueued
		return
	}

	// Locking the mutex to safely modify the active downloads count
	q.mu.Lock()
	q.ActiveDownloads++
	q.mu.Unlock()

	q.DownloadChannel <- struct{}{}
	d.Status = models.DownloadStatusDownloading
	log.Infof("Starting download: %s", d.URL)
	d.start(q)

	// Decrement active downloads after completion
	q.mu.Lock()
	q.ActiveDownloads--
	q.mu.Unlock()

	<-q.DownloadChannel
}
