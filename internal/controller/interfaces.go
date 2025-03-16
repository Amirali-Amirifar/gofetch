package controller

type DownloadManager interface {
	Create()
	Pause()
	Resume()
	Cancel()
}
