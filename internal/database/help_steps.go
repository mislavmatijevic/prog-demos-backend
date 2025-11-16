package database

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

func SaveHelpStep(helpStep *TaskHelpStep) error {
	result := Instance.db.Save(helpStep)
	return result.Error
}

func GetHelpStepByStepAndTaskId(step int, taskId int) *TaskHelpStep {
	var helpStep TaskHelpStep

	Instance.db.Where("step = ?", step).Where("id_task = ?", taskId).Find(&helpStep)

	if helpStep.ID == 0 {
		return nil
	}

	return &helpStep
}

func GetHelpStepCountByTaskId(taskId int) int64 {
	var count int64
	Instance.db.Model(&TaskHelpStep{}).Where("id_task = ?", taskId).Count(&count)
	return count
}

func MakeHelpStepAvailableForUser(userId int, taskId int, stepId int) error {
	var helpStep TaskHelpStep

	Instance.db.Where("step = ?", stepId).Where("id_task = ?", taskId).Find(&helpStep)

	if helpStep.ID == 0 {
		log.WithFields(log.Fields{"priority": "low", "taskId": taskId, "stepId": stepId}).Warn("Can't find TaskHelpStep for parameters.")
		return errors.New("Help step not found")
	}

	var availableHelpStep TaskAvailableHelpStep
	availableHelpStep.TaskHelpStep = &helpStep
	availableHelpStep.UserID = userId

	result := Instance.db.Save(&availableHelpStep)
	return result.Error
}

func GetAvailableHelpStepsForTaskByUser(userId int, taskId int) ([]TaskAvailableHelpStep, error) {
	var userAvailableHelpSteps []TaskAvailableHelpStep

	result := Instance.db.
		Joins("JOIN task_help_steps ON task_help_steps.id = task_available_help_steps.id_task_help_step").
		Where("task_available_help_steps.id_user = ? AND task_help_steps.id_task = ?", userId, taskId).
		Preload("TaskHelpStep").
		Find(&userAvailableHelpSteps)

	return userAvailableHelpSteps, result.Error
}
