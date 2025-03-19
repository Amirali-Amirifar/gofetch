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
)

type Download struct {
	models.Download
	contentLength int64
	contentType   string
	acceptRanges  bool
	rangesCount   int
	ranges        []int
}

func (d *Download) Create() {
	fileUrl := d.URL
	log.Infof("Creating download for URL: %s", fileUrl)

	// Make the request
	response, err := http.Get(fileUrl)
	if err != nil {
		log.Errorf("Failed to fetch URL: %s, error: %v", fileUrl, err)
		return
	}
	defer response.Body.Close()

	// Capture headers
	d.Headers = response.Header

	// Check status code
	if response.StatusCode != http.StatusOK {
		log.Errorf("Non-OK HTTP status: %d", response.StatusCode)
		return
	}

	if contentLength := d.Headers.Get("Content-Length"); contentLength != "" {
		d.contentLength, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			log.Errorf("Error parsing Content-Length: %v", err)
			d.contentLength = 0
		}
	} else {
		log.Warn("Missing Content-Length header")
	}

	d.acceptRanges = d.Headers.Get("Accept-Ranges") == "bytes"

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

	if d.FileName == "" {
		parsedURL, err := url.Parse(d.URL)
		if err == nil {
			segments := strings.Split(parsedURL.Path, "/")
			d.FileName = segments[len(segments)-1]
		}
	}

	if d.FileName == "" || !strings.Contains(d.FileName, ".") {
		if contentType := d.Headers.Get("Content-Type"); contentType != "" {
			ext, _ := mime.ExtensionsByType(contentType)
			if len(ext) > 0 {
				d.FileName = "download" + ext[0] // Default name with correct extension
			}
		}
	}

	log.Infof("Completed capturing initial info of %s, the data are %#v", d.URL, d)
	log.Infof("\nStarting the download process")

	d.start()
}

func (d *Download) start() {
	//if d.acceptRanges {
	//	d.startParallel()
	//} else {
	//	d.startSingleThread()
	//}

	d.startSingleThread()
}

func (d *Download) startParallel() {}

func (d *Download) startSingleThread() {
	if d.FileName == "" {
		log.Errorf("No filename provided")
		d.FileName = "GoFetch_Download.tmp"
	}

	// Ensure DefaultDownloadFolder is properly expanded
	downloadFolder := config.DefaultDownloadFolder
	if strings.HasPrefix(downloadFolder, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		downloadFolder = filepath.Join(homeDir, downloadFolder[2:])
	}

	// Set full file path
	d.FileName = filepath.Join(downloadFolder, d.FileName)

	// Ensure parent directories exist
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

	log.Infof("Created file %s", file.Name())

	// Get the HTTP response
	resp, err := http.Get(d.URL)
	if err != nil {
		log.Fatalf("Failed to download %s: %v", d.URL, err)
	}
	defer resp.Body.Close()

	// Write response body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatalf("Failed to write to file %s: %v", d.FileName, err)
	}

	log.Infof("Finished download process for %s", file.Name())
}
