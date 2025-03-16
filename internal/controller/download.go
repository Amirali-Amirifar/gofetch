package controller

import (
	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Download struct {
	models.Download
}

func (d *Download) Create() {
	log.Info("Creating download")
	targetLocation, err := os.Create("./dlFile.o")
	if err != nil {
		log.Errorf(err.Error())
		panic(err)
	}
	defer targetLocation.Close()

	request, err := http.Get(d.URL)
	if err != nil {
		log.Errorf(err.Error())
		panic(err)
	}

	log.Info(request.Header)
}
