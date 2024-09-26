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

func CheckIfUserAlreadySuccessfullyExecutedTask(userId int, taskId int) bool {
	var exists bool

	Instance.db.Model(&TaskExecution{}).
		Select("count(*) > 0").
		Where("id_user = @UserID AND id_task = @TaskID AND was_successful = @WasSuccessful",
			sql.Named("UserID", userId),
			sql.Named("TaskID", taskId),
			sql.Named("WasSuccessful", true),
		).Find(&exists)

	return exists
}

func SaveTaskExecution(taskExecution TaskExecution) (*TaskExecution, error) {
	result := Instance.db.Save(&taskExecution)
	return &taskExecution, result.Error
}
