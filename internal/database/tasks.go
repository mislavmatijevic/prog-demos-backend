package database

import (
	log "github.com/sirupsen/logrus"
)

func GetAllTasksPerTopic() []Topic {
	var topics []Topic

	result := Instance.db.Model(&Topic{}).Preload("Subtopics.Tasks").Find(&topics)

	if result.Error != nil {
		log.Error("Error fetching tasks: ", result.Error)
	}

	return topics
}
