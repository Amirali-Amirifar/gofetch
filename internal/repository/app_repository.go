package repository

import "github.com/Amirali-Amirifar/gofetch.git/internal/models"

type AppRepository interface {
	LoadAppState() (models.AppState, error)
	SaveAppState(appState models.AppState) error
}
