package database

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
