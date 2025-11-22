package taskexecution

import (
	"os"
	"strconv"

	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
)

var (
	CPP_FILE_NAME                      = ""
	TASK_RUNNER_TEMP_TASKS_DIRECTORY   = ""
	TASK_RUNNER_MAX_PARALLEL_AVAILABLE = 0
	TASK_RUNNER_MAX_CPUS_AVAILABLE     = 0.0
	TASK_RUNNER_MAX_RAM_MB_AVAILABLE   = 0
	TASK_RUNNER_MOUNT_NAME             = "" // Gets get overriden on dev! Either a docker volume name for prod or a local path for dev.
	BACKEND_TEMP_TASKS_DIRECTORY       = ""
)

func Initialize() {
	var err error
	CPP_FILE_NAME = os.Getenv("CPP_FILE_NAME")
	TASK_RUNNER_TEMP_TASKS_DIRECTORY = os.Getenv("TASK_RUNNER_TEMP_TASKS_DIRECTORY")
	TASK_RUNNER_MAX_PARALLEL_AVAILABLE, err = strconv.Atoi(os.Getenv("TASK_RUNNER_MAX_PARALLEL_AVAILABLE"))
	TASK_RUNNER_MAX_CPUS_AVAILABLE, err = strconv.ParseFloat(os.Getenv("TASK_RUNNER_MAX_CPUS_AVAILABLE"), 32)
	TASK_RUNNER_MAX_RAM_MB_AVAILABLE, err = strconv.Atoi(os.Getenv("TASK_RUNNER_MAX_RAM_MB_AVAILABLE"))
	if err != nil {
		log.WithFields(log.Fields{
			"TASK_RUNNER_MAX_PARALLEL_AVAILABLE": TASK_RUNNER_MAX_PARALLEL_AVAILABLE,
			"TASK_RUNNER_MAX_CPUS_AVAILABLE":     TASK_RUNNER_MAX_CPUS_AVAILABLE,
			"TASK_RUNNER_MAX_RAM_MB_AVAILABLE":   TASK_RUNNER_MAX_RAM_MB_AVAILABLE,
		}).WithError(err).Fatal("Failed to load task-runner available resources.")
	}

	if utils.IsProd() {
		TASK_RUNNER_MOUNT_NAME = os.Getenv("TASK_RUNNER_MOUNT_NAME")
		BACKEND_TEMP_TASKS_DIRECTORY = TASK_RUNNER_TEMP_TASKS_DIRECTORY
	} else {
		tempDirForTasks, err := os.MkdirTemp("", "prog-demos-backend-dev-temp-tasks-execution-dir-*")
		if err != nil {
			log.WithError(err).Fatal()
		}

		TASK_RUNNER_MOUNT_NAME = tempDirForTasks
		BACKEND_TEMP_TASKS_DIRECTORY = tempDirForTasks
	}
}
