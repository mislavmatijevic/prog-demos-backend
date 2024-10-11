package database

import (
	log "github.com/sirupsen/logrus"
)

func GetAllTasksPerTopic() []Topic {
	var topics []Topic

	result := Instance.db.Model(&Topic{}).Preload("Subtopics.Tasks").Find(&topics)

	if result.Error != nil {
		log.WithError(result.Error).WithFields(log.Fields{"priority": "high", "context": "tasks"}).Error("Couldn't fetch all tasks!")
	}

	return topics
}

func GetSingleFullTasks(taskId int) *FullTask {
	var task FullTask

	result := Instance.db.Preload("Subtopic").Preload("HelpSteps").First(&task, taskId)

	if result.Error != nil {
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
		log.WithError(result.Error).WithFields(log.Fields{"priority": "medium", "context": "tasks"}).Errorf("Couldn't fetch tests for task %d!", taskId)
		return nil
	}

	return tests
}

func SaveTask(task *FullTask) error {
	result := Instance.db.Save(task)
	return result.Error
}
