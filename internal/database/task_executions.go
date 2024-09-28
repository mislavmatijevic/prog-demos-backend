package database

import (
	"database/sql"
)

func GetRunningTaskExecutionForUserId(userId int) *TaskExecution {
	var taskExecution TaskExecution

	Instance.db.Where("id_user = @UserID AND is_finished = @IsFinished",
		sql.Named("UserID", userId),
		sql.Named("IsFinished", false),
	).Find(&TaskExecution{}).Scan(&taskExecution)

	if taskExecution.ID == 0 {
		return nil
	}

	return &taskExecution
}

func GetBestScoreExecutionOfUserForTask(userId int, taskId int) *TaskExecution {
	var foundSuccessfulExecution TaskExecution

	Instance.db.
		Where("id_user = @UserID AND id_task = @TaskID AND was_successful = @WasSuccessful AND best_score = @BestScore",
			sql.Named("UserID", userId),
			sql.Named("TaskID", taskId),
			sql.Named("WasSuccessful", true),
			sql.Named("BestScore", true),
		).Find(&TaskExecution{}).Scan(&foundSuccessfulExecution)

	if foundSuccessfulExecution.ID == 0 {
		return nil
	}

	return &foundSuccessfulExecution
}

func SaveTaskExecution(taskExecution TaskExecution) (*TaskExecution, error) {
	result := Instance.db.Save(&taskExecution)
	return &taskExecution, result.Error
}
