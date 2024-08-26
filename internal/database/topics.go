package database

import (
	log "github.com/sirupsen/logrus"
)

func GetAllTopicsWithSubtopics() []Topic {
	var topics []Topic

	result := Instance.db.Model(&Topic{}).Preload("Subtopics").Find(&topics)

	if result.Error != nil {
		log.Error("Error fetching topics: ", result.Error)
	}

	return topics
}
