package main

import (
	"os"

	"github.com/Amirali-Amirifar/gofetch.git/internal/config"
	"github.com/Amirali-Amirifar/gofetch.git/internal/repository/json"
	"github.com/Amirali-Amirifar/gofetch.git/internal/tui"
	log "github.com/sirupsen/logrus"
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(logFile)

	// Initialize the database
	_ = config.GetDB()
	defer func() {
		err := config.Close()
		if err != nil {
			log.Fatal("Failed to close DB:", err)
		}
	}()

	// Load or initialize the AppState
	state, err := json.LoadAppState()
	if err != nil {
		log.Fatalf("Failed to load app state: %v", err)
	}
	log.Println("App state loaded successfully")

	// Start the TUI
	program := tui.GetTui(state)
	log.Println("Starting TUI...")
	_, err = program.Run() // Capture both return values, ignore the model with _
	if err != nil {
		log.Fatalf("Failed to start TUI: %v", err)
	}
}
