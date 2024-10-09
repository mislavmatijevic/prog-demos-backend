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

	result := Instance.db.Preload("Subtopic").Preload("HelpSteps").First(&task, taskId)

	if result.Error != nil {
		log.Error("Error fetching task: ", result.Error)
		return nil
	}

	return &task
}

func CheckTaskExists(taskId int) bool {
	var task Task
	Instance.db.Find(&task, taskId)
	return task.ID != 0
}

func GetTestsForTask(taskId int) []TaskTest {
	var tests []TaskTest = make([]TaskTest, 0)

	result := Instance.db.Where("id_task=?", taskId).Find(&TaskTest{}).Scan(&tests)

	if result.Error != nil {
		log.Error("Error fetching task: ", result.Error)
		return nil
	}

	return tests
}

func SaveTask(task *FullTask) error {
	result := Instance.db.Save(task)
	return result.Error
}
