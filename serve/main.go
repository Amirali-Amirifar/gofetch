package main

import (
	"log"
	"net/http"
)

// main serves files from a directory with range requests enabled
// used for testing the downloader.
func main() {
	// Directory to serve files from
	fileDir := "./downloads"

	// Create a file server handler
	fs := http.FileServer(http.Dir(fileDir))

	// Custom handler to add Accept-Ranges support
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes") // Enable range requests
		fs.ServeHTTP(w, r)
	})

	port := "8080"
	log.Printf("Serving files from %s at http://localhost:%s\n", fileDir, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
