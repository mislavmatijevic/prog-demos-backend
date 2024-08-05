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

func GetSingleFullTasks(taskId int) *FullTask {
	var task FullTask

	result := Instance.db.First(&task, taskId)

	if result.Error != nil {
		log.Error("Error fetching task: ", result.Error)
		return nil
	}

	return &task
}
