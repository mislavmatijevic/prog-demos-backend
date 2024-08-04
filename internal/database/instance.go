package database

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type dbInstance struct {
	db *gorm.DB
}

var Instance dbInstance = dbInstance{}

func Initialize() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	var err error

	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName)

	Instance.db, err = gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	if err == nil {
		log.Info("Successfully connected to the database!")
	} else {
		log.Panicf("GORM failed to connect to the database! %s", err.Error())
	}

	err = Instance.db.AutoMigrate(&Topic{}, &Subtopic{}, &Video{}, &Task{}, &Test{})

	if err != nil {
		log.Warningf("Failed to do auto migration. Reason: %s", err)
	}
}

func GetAllVideosPerTopics() []Topic {
	var topics []Topic

	result := Instance.db.Model(&Topic{}).Preload("Subtopics.Videos").Find(&topics)

	if result.Error != nil {
		log.Error("Error fetching videos: ", result.Error)
	}

	return topics
}
