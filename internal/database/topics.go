package database

import (
	log "github.com/sirupsen/logrus"
)

func GetAllTopicsWithSubtopics() []Topic {
	var topics []Topic

	result := Instance.db.Model(&[]Topic{}).Preload("Subtopics").Find(&topics)

	if result.Error != nil {
		log.WithError(result.Error).WithFields(log.Fields{"priority": "medium", "context": "topics"}).Error("Couldn't fetch topics with subtopics!")
	}

	return topics
}

func GetSubtopicById(subtopicId int) *Subtopic {
	var subtopic Subtopic

	result := Instance.db.Find(&subtopic, subtopicId)

	if result.Error != nil || result.RowsAffected == 0 {
		return nil
	}

	return &subtopic
}
