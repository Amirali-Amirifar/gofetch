package models

type SwitchTabMsg struct {
	Direction string // "left" or "right"
}

type Queue struct {
	Name      string `json:"name"`
	Folder    string `json:"folder"`
	MaxDL     int    `json:"max_dl"`
	Speed     string `json:"speed"`
	TimeRange string `json:"time_range"`
}

type Download struct {
	URL      string `json:"url"`
	Queue    string `json:"queue"`
	Status   string `json:"status"`   // "Pending", "Downloading", "Paused", "Completed", "Failed"
	Progress int    `json:"progress"` // Percentage (0-100)
}

type AppState struct {
	Queues    []Queue    `json:"queues"`
	Downloads []Download `json:"downloads"`
}
