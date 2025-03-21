package config

const (
	DefaultQueueName        = "Default"
	DefaultDownloadFolder   = "~/Downloads/GoFetch/"
	DefaultDownloadSpeed    = 0 // 0 means unlimited
	DefaultMaxSimultaneous  = 3
	DefaultActiveTimeStart  = ""
	DefaultActiveTimeEnd    = ""
	DefaultMaxRetryAttempts = 3
	StateFile               = "state.json"
	databaseFile            = "sqlite3.db"
)
