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

func GetSubtopicById(subtopicId int) *Subtopic {
	var subtopic Subtopic

	result := Instance.db.Find(&subtopic, subtopicId)

	if result.Error != nil || result.RowsAffected == 0 {
		log.Error("Error fetching subtopic: ", result.Error)
		return nil
	}

	return &subtopic
}
