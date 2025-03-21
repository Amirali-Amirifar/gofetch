package controller

import (
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Amirali-Amirifar/gofetch.git/internal/config"
	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	log "github.com/sirupsen/logrus"
)

// Download represents a single download instance
// It embeds the Download model from the models package
type Download struct {
	models.Download
}

// Create initializes a new download and starts the download process if conditions allow
func (d *Download) Create(queue *QueueManager) {
	fileUrl := d.URL
	log.Infof("Creating download for URL: %s", fileUrl)
	d.Status = models.DownloadStatusQueued

	// Fetch the file metadata
	response, err := http.Get(fileUrl)
	if err != nil {
		log.Errorf("Failed to fetch URL: %s, error: %v", fileUrl, err)
		d.Status = models.DownloadStatusFailed
		return
	}
	defer response.Body.Close()
	d.Headers = response.Header

	if response.StatusCode != http.StatusOK {
		log.Errorf("Non-OK HTTP status: %d", response.StatusCode)
		return
	}

	// Extract content length if available
	if contentLength := d.Headers.Get("Content-Length"); contentLength != "" {
		d.ContentLength, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			log.Errorf("Error parsing Content-Length: %v", err)
			d.ContentLength = 0
		}
	} else {
		log.Warn("Missing Content-Length header")
	}

	d.AcceptRanges = d.Headers.Get("Accept-Ranges") == "bytes"
	d.FileName = d.extractFileName()

	log.Infof("Completed capturing initial info of %s, the data are %#v", d.URL, d)
	db := config.GetDB()
	if err := db.SaveDownload(d.Download); err != nil {
		log.Errorf("Failed to save download to database: %v", err)
		d.Status = models.DownloadStatusFailed
		return
	}

	queue.StartDownload(d)
}

// extractFileName attempts to determine the file name from the response headers or URL
func (d *Download) extractFileName() string {
	if contentDisposition := d.Headers.Get("Content-Disposition"); contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if filename, exists := params["filename"]; exists {
				return filename
			}
		}
	}

	parsedURL, err := url.Parse(d.URL)
	if err == nil {
		segments := strings.Split(parsedURL.Path, "/")
		return segments[len(segments)-1]
	}

	if contentType := d.Headers.Get("Content-Type"); contentType != "" {
		ext, _ := mime.ExtensionsByType(contentType)
		if len(ext) > 0 {
			return "download" + ext[0]
		}
	}

	return "GoFetch_Download.tmp"
}

// start initiates the actual file download process
func (d *Download) start(queue *QueueManager) {
	d.startSingleThread(queue)
}

// startSingleThread handles downloading the file in a single-threaded manner
func (d *Download) startSingleThread(queue *QueueManager) {
	// Determine the download folder
	downloadFolder := config.DefaultDownloadFolder
	if strings.HasPrefix(downloadFolder, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		downloadFolder = filepath.Join(homeDir, downloadFolder[2:])
	}
	d.FileName = filepath.Join(downloadFolder, d.FileName)
	err := os.MkdirAll(filepath.Dir(d.FileName), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory %s: %v", filepath.Dir(d.FileName), err)
	}

	// Create the file
	file, err := os.Create(d.FileName)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", d.FileName, err)
	}
	defer file.Close()

	// Fetch file data
	resp, err := http.Get(d.URL)
	if err != nil {
		log.Fatalf("Failed to download %s: %v", d.URL, err)
	}
	defer resp.Body.Close()

	// Apply bandwidth limiting
	limiter := queue.DownloadLimiter
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Generate a token each second
			limiter <- struct{}{}
		case <-limiter:
			// Proceed with download if token is available
			_, err := io.Copy(file, resp.Body)
			if err != nil {
				log.Fatalf("Failed to write to file %s: %v", d.FileName, err)
			}
			log.Infof("Finished download process for %s", file.Name())
			return
		}
	}
}
