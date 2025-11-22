package taskexecution

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	log "github.com/sirupsen/logrus"
)

var (
	ErrUserHasRunningTasks   = errors.New("user has running task")
	ErrTaskExecutionStartErr = errors.New("couldn't mark execution as started")
	ErrTempFileCreationErr   = errors.New("couldn't create temp file")
	ErrNoTests               = errors.New("no tests")
)

type TaskExecutionRequestInfo struct {
	Code   string
	UserId int
	TaskId int
}

func PrepareTaskExecution(requestInfo TaskExecutionRequestInfo) (*TaskExecutionData, error) {
	var newData = &TaskExecutionData{}
	var err error

	alreadyHasRunningTask := checkUserHasRunningTasks(requestInfo.UserId)
	if alreadyHasRunningTask {
		return newData, ErrUserHasRunningTasks
	}

	newData.InitializedTaskExecution, err = markTaskExecutionStartForUserId(requestInfo.TaskId, requestInfo.UserId, requestInfo.Code)
	if err != nil {
		log.WithError(err).Debug("Failed to mark task execution as started.")
		return newData, ErrTaskExecutionStartErr
	}

	newData.File, err = createCppFileInNewTempDirectory(requestInfo.Code)
	if err != nil {
		log.WithError(err).Debug("Failed to create temp file.")
		return newData, ErrTempFileCreationErr
	}
	newData.tempFolderPath = path.Dir(newData.File.Name())

	newData.Tests = database.GetTestsForTask(requestInfo.TaskId)
	if len(newData.Tests) == 0 {
		return newData, ErrNoTests
	}

	return newData, err
}

func checkUserHasRunningTasks(userId int) bool {
	currentlyRunningTaskExecution := database.GetRunningTaskExecutionForUserId(userId)
	return currentlyRunningTaskExecution != nil
}

func markTaskExecutionStartForUserId(taskId, userId int, code string) (*database.TaskExecution, error) {
	var taskExecution database.TaskExecution = database.TaskExecution{
		InitiatorID:   userId,
		TaskID:        taskId,
		StartedAt:     time.Now(),
		IsFinished:    false,
		SubmittedCode: code,
		WasSuccessful: false,
	}

	return database.SaveTaskExecution(taskExecution)
}

func createCppFileInNewTempDirectory(cppCode string) (*os.File, error) {
	createdTempPath, err := os.MkdirTemp(BACKEND_TEMP_TASKS_DIRECTORY, "temp_cpp_solutions_*")
	if err != nil {
		return nil, err
	}

	createdTempCppFile, err := createFile(createdTempPath, CPP_FILE_NAME, cppCode)
	if err != nil {
		return nil, err
	}

	return createdTempCppFile, nil
}

// No file gets created if I fail.
func createFile(path string, name string, contents string) (*os.File, error) {
	newFile, err := os.Create(filepath.Join(path, name))
	if err != nil {
		return nil, err
	}

	newFile.Chmod(0640)

	_, err = newFile.Write([]byte(contents))
	if err != nil {
		os.Remove(newFile.Name())
	}

	return newFile, err
}
