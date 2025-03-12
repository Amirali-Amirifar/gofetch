package main

import (
	"log"

	"github.com/Amirali-Amirifar/gofetch.git/internal/tui"
)

func main() {
	// Load or initialize the AppState
	state, err := tui.LoadAppState()
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
