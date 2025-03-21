package controller

import (
	"github.com/Amirali-Amirifar/gofetch.git/internal/config"
	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	log "github.com/sirupsen/logrus"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Download struct {
	models.Download
	// Exported fields for progress tracking.
	ContentLength   int64     // Total size in bytes.
	CurrentProgress int64     // Bytes downloaded so far.
	StartTime       time.Time // When the download started.

	// Internal fields.
	contentType  string
	acceptRanges bool
	rangesCount  int
	ranges       []int
}

// Create gathers initial info and kicks off the download.
func (d *Download) Create() {
	fileUrl := d.URL
	log.Infof("Creating download for URL: %s", fileUrl)

	// Make a HEAD request to obtain headers.
	response, err := http.Get(fileUrl)
	if err != nil {
		log.Errorf("Failed to fetch URL: %s, error: %v", fileUrl, err)
		return
	}
	defer response.Body.Close()

	d.Headers = response.Header
	if response.StatusCode != http.StatusOK {
		log.Errorf("Non-OK HTTP status: %d", response.StatusCode)
		return
	}

	// Extract and parse Content-Length.
	if cl := d.Headers.Get("Content-Length"); cl != "" {
		d.ContentLength, err = strconv.ParseInt(cl, 10, 64)
		if err != nil {
			log.Errorf("Error parsing Content-Length: %v", err)
			d.ContentLength = 0
		}
	} else {
		log.Warn("Missing Content-Length header")
	}

	d.acceptRanges = d.Headers.Get("Accept-Ranges") == "bytes"

	// Try to extract filename from Content-Disposition header.
	if contentDisposition := d.Headers.Get("Content-Disposition"); contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if filename, exists := params["filename"]; exists {
				d.FileName = filename
				log.Infof("Extracted filename: %s", filename)
			}
		} else {
			log.Warnf("Failed to parse Content-Disposition: %v", err)
		}
	}

	// Fallback to the last segment of URL path.
	if d.FileName == "" {
		parsedURL, err := url.Parse(d.URL)
		if err == nil {
			segments := strings.Split(parsedURL.Path, "/")
			d.FileName = segments[len(segments)-1]
		}
	}

	// Final fallback to a default filename with inferred extension.
	if d.FileName == "" || !strings.Contains(d.FileName, ".") {
		if contentType := d.Headers.Get("Content-Type"); contentType != "" {
			ext, _ := mime.ExtensionsByType(contentType)
			if len(ext) > 0 {
				d.FileName = "download" + ext[0]
			}
		}
	}

	log.Infof("Completed capturing initial info of %s, details: %#v", d.URL, d)
	log.Infof("Starting the download process")
	d.start()
}

func (d *Download) start() {
	// Right now, we support only single-threaded download.
	d.startSingleThread()
}

func (d *Download) startSingleThread() {
	if d.FileName == "" {
		log.Errorf("No filename provided")
		d.FileName = "GoFetch_Download.tmp"
	}

	// Expand download folder (e.g., handling '~' prefix).
	downloadFolder := config.DefaultDownloadFolder
	if strings.HasPrefix(downloadFolder, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		downloadFolder = filepath.Join(homeDir, downloadFolder[2:])
	}

	// Set the full path.
	d.FileName = filepath.Join(downloadFolder, d.FileName)
	if err := os.MkdirAll(filepath.Dir(d.FileName), os.ModePerm); err != nil {
		log.Fatalf("Failed to create directory %s: %v", filepath.Dir(d.FileName), err)
	}

	file, err := os.Create(d.FileName)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", d.FileName, err)
	}
	defer file.Close()
	log.Infof("Created file %s", file.Name())

	// Make the actual GET request.
	resp, err := http.Get(d.URL)
	if err != nil {
		log.Fatalf("Failed to download %s: %v", d.URL, err)
	}
	defer resp.Body.Close()

	// Record the start time before reading.
	d.StartTime = time.Now()
	var totalWritten int64 = 0
	buf := make([]byte, 32*1024) // 32KB buffer

	// Read from network and write to file in chunks.
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			written, err2 := file.Write(buf[:n])
			if err2 != nil {
				log.Fatalf("Error writing to file %s: %v", d.FileName, err2)
			}
			totalWritten += int64(written)
			d.CurrentProgress = totalWritten
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading response body: %v", err)
		}
	}

	log.Infof("Finished download process for %s, total bytes written: %d", file.Name(), totalWritten)
}
