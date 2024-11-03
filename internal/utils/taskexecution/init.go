package taskexecution

import (
	"os"

	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

var (
	CPP_FILE_NAME                = ""
	TASKS_VOLUME_NAME            = ""
	MOUNTED_TEMP_TASKS_DIRECTORY = ""
	LOCAL_TEMP_TASKS_DIRECTORY   = ""
)

func Initialize() {
	CPP_FILE_NAME = os.Getenv("CPP_FILE_NAME")
	TASKS_VOLUME_NAME = os.Getenv("TASKS_VOLUME_NAME")
	MOUNTED_TEMP_TASKS_DIRECTORY = os.Getenv("MOUNTED_TEMP_TASKS_DIRECTORY")
	if utils.IsProd() {
		LOCAL_TEMP_TASKS_DIRECTORY = MOUNTED_TEMP_TASKS_DIRECTORY
	}
}
