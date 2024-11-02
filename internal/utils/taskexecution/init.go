package taskexecution

import (
	"os"
)

var (
	CPP_FILE_NAME                    = ""
	TASKS_VOLUME_NAME                = ""
	LOCAL_TEMP_TASKS_DIRECTORY       = ""
	TASK_RUNNER_TEMP_TASKS_DIRECTORY = ""
)

func Initialize() {
	CPP_FILE_NAME = os.Getenv("CPP_FILE_NAME")
	TASKS_VOLUME_NAME = os.Getenv("TASKS_VOLUME_NAME")
	LOCAL_TEMP_TASKS_DIRECTORY = os.Getenv("LOCAL_TEMP_TASKS_DIRECTORY")
	TASK_RUNNER_TEMP_TASKS_DIRECTORY = os.Getenv("TASK_RUNNER_TEMP_TASKS_DIRECTORY")
}
