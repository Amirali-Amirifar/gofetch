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
	URL      string `json":"url"`
	Queue    string `json:"queue"`
	FileName string `json:"file_name"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

type AppState struct {
	Queues    []Queue    `json:"queues"`
	Downloads []Download `json:"downloads"`
}
