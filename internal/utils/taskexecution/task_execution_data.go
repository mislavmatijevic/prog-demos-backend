package taskexecution

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/lizard"
	log "github.com/sirupsen/logrus"
)

type TaskExecutionData struct {
	File                     *os.File
	InitializedTaskExecution *database.TaskExecution
	Tests                    []database.TaskTest
	CreatedInputFile         *os.File
	tempFolderPath           string
}

func (executionData *TaskExecutionData) ReadAllArtefactFiles() ([]byte, error) {
	artefactsFilePath := filepath.Join(executionData.tempFolderPath, "artefacts.txt")
	bytes, errMain := os.ReadFile(artefactsFilePath)
	var stringOutput = string(bytes)
	var fileNames = strings.Split(stringOutput, "\n")
	sort.Strings(fileNames)

	var allBytes []byte = make([]byte, 0)

	for _, relativeFileName := range fileNames {
		if len(relativeFileName) == 0 {
			continue
		}

		var actualFileName = filepath.Base(relativeFileName)
		fullFilePath := filepath.Join(executionData.tempFolderPath, actualFileName)

		currentFileBytes, err := os.ReadFile(fullFilePath)
		if err != nil {
			errMain = err
			break
		}
		allBytes = append(allBytes, currentFileBytes...)
	}

	return allBytes, errMain
}

func (executionData *TaskExecutionData) CreateInputFile(test database.TaskTest) (err error) {
	var inputFileName = getFilenameBasedOnTest(STDIN_FILENAME_PREFIX, test.ID)
	executionData.CreatedInputFile, err = createFile(executionData.tempFolderPath, inputFileName, test.Input)
	return err
}

func (executionData *TaskExecutionData) ReadOutputFile(test database.TaskTest) (contents []byte, err error) {
	return executionData.readContentsFromTestOutputFile(STDOUT_FILENAME_PREFIX, test.ID)
}

func (executionData *TaskExecutionData) ReadSha256FromArtefactFile(test database.TaskTest) (contents []byte, err error) {
	return executionData.readContentsFromTestOutputFile(ARTEFACTS_FILENAME_PREFIX, test.ID)
}

func (executionData *TaskExecutionData) readContentsFromTestOutputFile(filenamePrefix string, testId int) (contents []byte, err error) {
	var expectedOutputFileName = getFilenameBasedOnTest(filenamePrefix, testId)
	var testOutputPath = filepath.Join(executionData.tempFolderPath, expectedOutputFileName)
	return os.ReadFile(testOutputPath)
}

func (executionData *TaskExecutionData) ReadErrorFile() (contents []byte, err error) {
	var testOutputPath = filepath.Join(executionData.tempFolderPath, ERROR_FILENAME)
	return os.ReadFile(testOutputPath)
}

func getFilenameBasedOnTest(prefix string, testId int) string {
	return fmt.Sprintf("%s%d.txt", prefix, testId)
}

func (executionData *TaskExecutionData) CleanupTempFolder() {
	os.RemoveAll(executionData.tempFolderPath)
}

// Persists info that task execution failed and clears temp folder.
func (executionData *TaskExecutionData) SetTaskExecutionStatusFailed() {
	executionData.InitializedTaskExecution.WasSuccessful = false
	saveFinishedTaskExecution(executionData.InitializedTaskExecution)
	executionData.CleanupTempFolder()
}

func (data *TaskExecutionData) SetTaskExecutionStatusSucceeded(solvedTask *database.FullTask, score lizard.CodeScore) {
	var taskExecution *database.TaskExecution = data.InitializedTaskExecution

	taskExecution.WasSuccessful = true
	taskExecution.CodeScore = &score

	var previousBestScoreExecutionFromThisUserForThisTask = database.GetBestScoreExecutionForUserAndTask(
		taskExecution.InitiatorID,
		taskExecution.TaskID,
	)

	var newUsersBestScore = false
	var newTaskAllTimeBestScore = false

	if previousBestScoreExecutionFromThisUserForThisTask == nil {
		newUsersBestScore = true
	} else {
		if taskExecution.CodeScore.HasBetterScoreThan(previousBestScoreExecutionFromThisUserForThisTask.CodeScore) {
			subtractLastScore(solvedTask, *previousBestScoreExecutionFromThisUserForThisTask.CodeScore)
			newUsersBestScore = true
		}
	}

	if newUsersBestScore {
		taskExecution.BestScore = true
		increaseAverageScoreOnTaskItself(solvedTask, score)
		increaseUserTotalScore(taskExecution)

		if previousBestScoreExecutionFromThisUserForThisTask != nil {
			previousBestScoreExecutionFromThisUserForThisTask.BestScore = false
			database.SaveTaskExecution(*previousBestScoreExecutionFromThisUserForThisTask)
			decreaseUserTotalScore(previousBestScoreExecutionFromThisUserForThisTask)
		}
	}

	if taskExecution.CodeScore.HasBetterScoreThan(&solvedTask.AllTimeBestScore) {
		solvedTask.AllTimeBestScore = *taskExecution.CodeScore
		newTaskAllTimeBestScore = true
	}

	if newUsersBestScore || newTaskAllTimeBestScore {
		database.SaveTask(solvedTask)
	}

	saveFinishedTaskExecution(taskExecution)
}

func subtractLastScore(fullTask *database.FullTask, score lizard.CodeScore) {
	scoresCountSoFar, tokensSum, scoresSum, complexitySum := getTaskScoreSums(fullTask)

	tokensSum -= score.Tokens
	scoresSum -= score.TotalScore
	complexitySum -= score.Complexity
	scoresCountSoFar--

	setAverageTaskScore(fullTask, scoresCountSoFar, tokensSum, scoresSum, complexitySum)
}

func increaseAverageScoreOnTaskItself(fullTask *database.FullTask, score lizard.CodeScore) {
	scoresCountSoFar, tokensSum, scoresSum, complexitySum := getTaskScoreSums(fullTask)

	tokensSum += score.Tokens
	scoresSum += score.TotalScore
	complexitySum += score.Complexity
	scoresCountSoFar++

	setAverageTaskScore(fullTask, scoresCountSoFar, tokensSum, scoresSum, complexitySum)
}

func getTaskScoreSums(task *database.FullTask) (int, int, int, int) {
	var scoresCountSoFar = task.ScoresCount
	var tokensSum = task.AverageScore.Tokens * scoresCountSoFar
	var scoresSum = task.AverageScore.TotalScore * scoresCountSoFar
	var complexitySum = task.AverageScore.Complexity * scoresCountSoFar
	return scoresCountSoFar, tokensSum, scoresSum, complexitySum
}

func setAverageTaskScore(task *database.FullTask, newScoresCount int, tokensSum int, scoresSum int, complexitySum int) {
	task.ScoresCount = newScoresCount
	if newScoresCount == 0 {
		task.AverageScore.Tokens = 0
		task.AverageScore.TotalScore = 0
		task.AverageScore.Complexity = 0
	} else {
		task.AverageScore.Tokens = tokensSum / newScoresCount
		task.AverageScore.TotalScore = utils.FloatToInt(float32(scoresSum) / float32(newScoresCount))
		task.AverageScore.Complexity = complexitySum / newScoresCount
	}
}

func increaseUserTotalScore(taskExecution *database.TaskExecution) {
	user, err := database.GetUserById(taskExecution.InitiatorID)
	if err != nil {
		logFailedToSaveUserTotalScore(err, taskExecution)
		return
	}

	user.TotalScore += taskExecution.TotalScore

	err = database.SaveUser(*user)
	if err != nil {
		logFailedToSaveUserTotalScore(err, taskExecution)
	}
}

func decreaseUserTotalScore(taskExecution *database.TaskExecution) {
	user, err := database.GetUserById(taskExecution.InitiatorID)
	if err != nil {
		logFailedToSaveUserTotalScore(err, taskExecution)
		return
	}

	user.TotalScore -= taskExecution.TotalScore

	err = database.SaveUser(*user)
	if err != nil {
		logFailedToSaveUserTotalScore(err, taskExecution)
	}
}

func logFailedToSaveUserTotalScore(err error, taskExecution *database.TaskExecution) {
	log.WithError(err).WithFields(
		log.Fields{"priority": "high",
			"context": "task_execution",
			"task_id": taskExecution.TaskID,
			"user_id": taskExecution.InitiatorID,
			"scored":  taskExecution.TotalScore,
		},
	).Error("Failed to save user total score!")
}

func saveFinishedTaskExecution(taskExecution *database.TaskExecution) {
	taskExecution.IsFinished = true
	taskExecution.FinishedAt = time.Now()
	database.SaveTaskExecution(*taskExecution)
}
