package config

import (
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

const defaultLocalDSN = "host=localhost user=postgres password=postgres dbname=budgeting_api port=5432 sslmode=disable TimeZone=Asia/Jakarta"

func ConnectDB() {
	onlineDSN := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	localDSN := strings.TrimSpace(os.Getenv("LOCAL_DATABASE_URL"))
	if localDSN == "" {
		localDSN = strings.TrimSpace(os.Getenv("DATABASE_URL_LOCAL"))
	}
	if localDSN == "" {
		localDSN = defaultLocalDSN
	}

	var db *gorm.DB
	var err error

	if onlineDSN != "" {
		db, err = openPostgres(onlineDSN)
		if err == nil {
			DB = db
			log.Println("Online database connected")
			return
		}

		log.Printf("Online database connection failed, trying local database: %v", err)
	} else {
		log.Println("DATABASE_URL is empty, trying local database")
	}

	db, err = openPostgres(localDSN)
	if err != nil {
		log.Fatal("Failed to connect local database:", err)
	}

	DB = db
	log.Println("Local database connected")
}

func openPostgres(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
