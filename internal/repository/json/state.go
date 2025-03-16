package json

import (
	"encoding/json"
	"fmt"
	"github.com/Amirali-Amirifar/gofetch.git/internal/config"
	"os"

	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
)

const stateFile = config.StateFile

func LoadAppState() (models.AppState, error) {
	var state models.AppState

	// Check if the file exists
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		// Create default state if file doesn't exist
		state = models.AppState{
			Queues: []models.Queue{
				{Name: config.DefaultQueueName, Folder: config.DefaultDownloadFolder, MaxDL: 3, Speed: config.DefaultDownloadSpeed, TimeRange: "24/7"},
			},
			Downloads: []models.Download{},
		}
		// Save the default state to the file
		if err := SaveAppState(state); err != nil {
			return state, fmt.Errorf("error saving default state: %w", err)
		}
		return state, nil
	}

	// Load state from file
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return state, fmt.Errorf("error reading state file: %w", err)
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("error unmarshaling state: %w", err)
	}
	return state, nil
}

// SaveAppState saves the entire application state to a single file.
func SaveAppState(state models.AppState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling state: %w", err)
	}

	// Directly check the error returned by os.WriteFile
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("error writing state file: %w", err)
	}

	return nil
}
