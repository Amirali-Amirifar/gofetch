package models

type SwitchTabMsg struct {
	Direction string // "left" or "right"
}

type AppState struct {
	Queues    []Queue    `json:"queues"`
	Downloads []Download `json:"downloads"`
}
