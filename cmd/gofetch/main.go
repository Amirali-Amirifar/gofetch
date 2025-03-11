package main

import (
	"github.com/Amirali-Amirifar/gofetch.git/internal/tui"
)

func main() {
	// Load or initialize the AppState
	state, err := tui.LoadAppState()
	if err != nil {
		panic(err)
	}

	// Start the TUI
	program := tui.GetTui(state)
	if err := program.Start(); err != nil {
		panic(err)
	}
}
