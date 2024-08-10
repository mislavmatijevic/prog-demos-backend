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

	result := Instance.db.Preload("Subtopic").First(&task, taskId)

	if result.Error != nil {
		log.Error("Error fetching task: ", result.Error)
		return nil
	}

	return &task
}

func CheckTaskExists(taskId int) bool {
	var task BasicTask
	Instance.db.Find(&task, taskId)
	return task.ID != 0
}

func GetTestsForTask(taskId int) []Test {
	var tests []Test = make([]Test, 0)

	result := Instance.db.Where("id_task=?", taskId).Find(&Test{}).Scan(&tests)

	if result.Error != nil {
		log.Error("Error fetching task: ", result.Error)
		return nil
	}

	return tests
}
