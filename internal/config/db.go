package config

import (
	"github.com/Amirali-Amirifar/gofetch.git/internal/repository/sqliteDb"
	log "github.com/sirupsen/logrus"
)

var DB *sqliteDb.SQLiteRepository

func GetDB() *sqliteDb.SQLiteRepository {
	if DB == nil {
		instance, err := sqliteDb.New(databaseFile)
		if err != nil {
			log.Fatalf("Error connecting to sqlite3 database: %s", err)
		}
		DB = instance
	}
	return DB
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
