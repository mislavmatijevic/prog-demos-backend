package database

func SaveHelpStep(helpStep *TaskHelpStep) error {
	result := Instance.db.Save(helpStep)
	return result.Error
}

func GetHelpStepByStepAndTaskId(step int, task int) *TaskHelpStep {
	var helpStep TaskHelpStep

	Instance.db.Where("step = ?", step).Where("id_task = ?", task).Find(&helpStep)

	if helpStep.ID == 0 {
		return nil
	}

	return &helpStep
}
