package controller

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Amirali-Amirifar/gofetch.git/internal/config"
	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	log "github.com/sirupsen/logrus"
)

type Download struct {
	models.Download
}

var db = config.GetDB()

// PauseDownload sets the download state to paused.
func (d *Download) PauseDownload() {
	if d.Status == models.DownloadStatusDownloading {
		d.updateStatus(models.DownloadStatusPaused)
		log.Infof("Download paused for %s", d.URL)
	}
}

// ResumeDownload resumes a paused download.
func (d *Download) ResumeDownload() {
	if d.Status == models.DownloadStatusPaused {
		d.updateStatus(models.DownloadStatusDownloading)
		log.Infof("Download resumed for %s", d.URL)
	}
}

// CancelDownload cancels the download.
func (d *Download) CancelDownload() {
	if d.Status != models.DownloadStatusCanceled && d.Status != models.DownloadStatusCompleted {
		d.updateStatus(models.DownloadStatusCanceled)
		if d.CancelChan != nil {
			close(d.CancelChan)
		}
		log.Infof("Download canceled for %s", d.URL)
	}
}

// Create gathers initial info (headers, inferred filename, etc.) and kicks off the download.
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
		d.ContentLength, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			log.Errorf("Error parsing Content-Length: %v", err)
			d.ContentLength = 0
		}
	} else {
		log.Warn("Missing Content-Length header")
	}

	d.AcceptRanges = d.Headers.Get("Accept-Ranges") == "bytes"

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

	log.Infof("Completed capturing initial info of %s, details: %#v", d.URL, d)
	log.Infof("Starting the download process")

	d.CancelChan = make(chan struct{})
	d.Status = models.DownloadStatusQueued

	err = db.AddNewDownload(&d.Download)
	if err != nil {
		log.Errorf("Failed to save download: %v", err)
	}
	d.start()
}

func (d *Download) start() {
	// Define a threshold for multipart downloads (e.g., 10 MB).
	d.updateStatus(models.DownloadStatusDownloading)
	err := db.UpdateDownload(&d.Download)
	if err != nil {
		log.Errorf("Failed to update download: %v", err)
		return
	}
	const multiPartThreshold int64 = 10 * 1024 * 1024
	if d.AcceptRanges && d.ContentLength > multiPartThreshold {
		log.Infof("Server supports multi-part and file size (%d bytes) exceeds threshold. Starting parallel download.", d.ContentLength)
		d.startParallel()
	} else {
		log.Infof("Starting single-threaded download")
		d.startSingleThread()
	}
}
func uniqueFileName(filePath string) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath
	}

	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	nameOnly := strings.TrimSuffix(base, ext)

	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s(%d)%s", nameOnly, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

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

	// Record the start time and update status.
	d.StartTime = time.Now()
	d.updateStatus(models.DownloadStatusDownloading)

	var totalWritten int64 = 0
	buf := make([]byte, 32*1024) // 32KB buffer

	for {

		select {
		case <-d.CancelChan:
			log.Infof("Download canceled for %s", d.URL)
			return
		default:
		}

		if d.Status == models.DownloadStatusPaused {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			written, err2 := file.Write(buf[:n])
			if err2 != nil {
				log.Fatalf("Error writing to file %s: %v", d.FileName, err2)
			}
			totalWritten += int64(written)
			d.updateProgress(totalWritten)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading response body: %v", err)
		}
	}

	if d.Status != models.DownloadStatusCanceled {
		d.updateStatus(models.DownloadStatusCompleted)
		log.Infof("Finished download process for %s, total bytes written: %d", file.Name(), totalWritten)
	}
}

func (d *Download) startParallel() {
	if d.FileName == "" {
		log.Errorf("No filename provided")
		d.FileName = "GoFetch_Download.tmp"
	}

	// Expand the default download folder.
	downloadFolder := config.DefaultDownloadFolder
	if strings.HasPrefix(downloadFolder, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		downloadFolder = filepath.Join(homeDir, downloadFolder[2:])
	}

	if !filepath.IsAbs(d.FileName) {
		d.FileName = filepath.Join(downloadFolder, d.FileName)
	}

	if err := os.MkdirAll(filepath.Dir(d.FileName), os.ModePerm); err != nil {
		log.Fatalf("Failed to create directory %s: %v", filepath.Dir(d.FileName), err)
	}

	// Ensure a unique filename.
	d.FileName = uniqueFileName(d.FileName)

	// We use a 2MB chunk size.
	const chunkSize int64 = 2 * 1024 * 1024
	numParts := int(d.ContentLength / chunkSize)
	if d.ContentLength%chunkSize != 0 {
		numParts++
	}
	// Limit the number of parts to the maximum concurrent downloads defined in configuration.
	maxConc := config.MaxConcurrentDownloads
	if numParts > maxConc {
		numParts = maxConc
	}
	log.Infof("Downloading in %d parts", numParts)

	type byteRange struct {
		start int64
		end   int64
	}
	var ranges []byteRange
	partSize := d.ContentLength / int64(numParts)
	var startByte int64 = 0
	for i := 0; i < numParts; i++ {
		var endByte int64
		if i == numParts-1 {
			endByte = d.ContentLength - 1
		} else {
			endByte = startByte + partSize - 1
		}
		ranges = append(ranges, byteRange{start: startByte, end: endByte})
		startByte = endByte + 1
	}

	// Prepare temporary filenames for each part.
	tempFiles := make([]string, numParts)
	var wg sync.WaitGroup
	progressChan := make(chan int64)
	errorChan := make(chan error, numParts)

	// Record start time and update status.
	d.StartTime = time.Now()
	d.updateStatus(models.DownloadStatusDownloading)

	// Start downloading each part concurrently.
	for i, r := range ranges {
		wg.Add(1)
		tempFileName := fmt.Sprintf("%s.part_%d", d.FileName, i)
		tempFiles[i] = tempFileName
		go d.downloadPart(i, r.start, r.end, tempFileName, progressChan, &wg, errorChan)
	}
	totalProgress := int64(0)
	// Aggregate progress from all parts.
	go func() {
		for p := range progressChan {
			totalProgress += p
			d.updateProgress(totalProgress)
		}
	}()

	wg.Wait()
	close(progressChan)

	// Check for any errors.
	select {
	case err := <-errorChan:
		if err != nil {
			log.Errorf("Error during parallel download: %v", err)
			return
		}
	default:
		// No error.
	}

	// Merge all part files into the final file.
	finalFile, err := os.Create(d.FileName)
	if err != nil {
		log.Fatalf("Failed to create final file %s: %v", d.FileName, err)
	}
	defer finalFile.Close()

	for i := 0; i < numParts; i++ {
		partFile, err := os.Open(tempFiles[i])
		if err != nil {
			log.Fatalf("Failed to open part file %s: %v", tempFiles[i], err)
		}
		_, err = io.Copy(finalFile, partFile)
		partFile.Close()
		if err != nil {
			log.Fatalf("Failed to merge part file %s: %v", tempFiles[i], err)
		}
		// Remove temporary part file.
		os.Remove(tempFiles[i])
	}

	if d.Status != models.DownloadStatusCanceled {
		d.updateStatus(models.DownloadStatusCompleted)
		log.Infof("Finished parallel download for %s", d.FileName)
	}
}

func (d *Download) downloadPart(index int, startByte, endByte int64, tempFileName string, progressChan chan<- int64, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()

	req, err := http.NewRequest("GET", d.URL, nil)
	if err != nil {
		errorChan <- fmt.Errorf("part %d: %v", index, err)
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startByte, endByte))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errorChan <- fmt.Errorf("part %d: %v", index, err)
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(tempFileName)
	if err != nil {
		errorChan <- fmt.Errorf("part %d: %v", index, err)
		return
	}
	defer file.Close()

	buf := make([]byte, 32*1024) // 32KB chunks
	for {

		select {
		case <-d.CancelChan:
			errorChan <- fmt.Errorf("part %d: canceled", index)
			return
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			written, err2 := file.Write(buf[:n])
			if err2 != nil {
				errorChan <- fmt.Errorf("part %d: %v", index, err2)
				return
			}
			progressChan <- int64(written)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			errorChan <- fmt.Errorf("part %d: %v", index, err)
			return
		}
	}
}

func (d *Download) updateStatus(status models.DownloadStatus) {
	d.Status = status
	func() {
		err := db.UpdateDownload(&d.Download)
		if err != nil {
			log.Errorf("Failed to update download status: %v", err)
		}
	}()
}

func (d *Download) updateProgress(totalWritten int64) {
	d.CurrentProgress = totalWritten
	if d.ContentLength != 0 {
		d.Progress = int(totalWritten/d.ContentLength) * 100
	}
}
