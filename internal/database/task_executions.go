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

func SaveTaskExecution(taskExecution TaskExecution) (*TaskExecution, error) {
	result := Instance.db.Save(&taskExecution)
	return &taskExecution, result.Error
}
