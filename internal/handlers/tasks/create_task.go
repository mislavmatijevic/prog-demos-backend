package tasks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

type taskTestBody struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expectedOutput"`
	ArtefactSHA256 string `json:"artefactSha256,omitempty"`
}

type taskHelpBody struct {
	Step       int    `json:"step"`
	HelperCode string `json:"helperCode,omitempty"`
	HelperText string `json:"helperText,omitempty"`
}

type newTaskRequestBody struct {
	SubtopicID         int            `json:"idSubtopic"`
	Name               string         `json:"name"`
	Complexity         string         `json:"complexity"`
	Input              string         `json:"input"`
	Output             string         `json:"output"`
	InputOutputExample string         `json:"inputOutputExample"`
	IsFinalBoss        bool           `json:"isFinalBoss"`
	SolutionCode       string         `json:"solutionCode,omitempty"`
	Tests              []taskTestBody `json:"tests"`
	HelpSteps          []taskHelpBody `json:"helpSteps"`
}

type responseBody struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NewTaskId int    `json:"newTaskId,omitempty"`
}

func (body newTaskRequestBody) mapToEntity() (newFullTaskEntity *database.FullTask) {
	return &database.FullTask{
		SubtopicID:         body.SubtopicID,
		Name:               body.Name,
		Complexity:         body.Complexity,
		Input:              body.Input,
		Output:             body.Output,
		InputOutputExample: body.InputOutputExample,
		IsFinalBoss:        body.IsFinalBoss,
		SolutionCode:       body.SolutionCode,
		ID:                 0,
		CreatorID:          0,
		Subtopic:           nil,
		Tests:              nil,
		Creator:            nil,
	}

}

func createTask(w http.ResponseWriter, r *http.Request) {
	newTask, err := getNewTaskFromBody(r)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	err = validateNewTask(*newTask)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	err = checkForValidTests(*newTask)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	var taskEntity *database.FullTask = newTask.mapToEntity()

	err = attachCreatorIdToTask(r, taskEntity)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	attachTestsToTask(newTask.Tests, taskEntity)

	if !newTask.IsFinalBoss {
		err = checkForValidHelpSteps(*newTask)
		if err != nil {
			api.RequestErrorHandlerCustomMsg(w, err.Error())
			return
		}
		attachHelpStepsToTask(newTask.HelpSteps, taskEntity)
	}

	err = storeTaskInDatabase(taskEntity)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(responseBody{Success: true, Message: fmt.Sprintf("Created new task with ID %v", taskEntity.ID)})
}

func getNewTaskFromBody(r *http.Request) (*newTaskRequestBody, error) {
	if r.Body == nil {
		return nil, errors.New("body is missing task's data")
	}

	var requestBody newTaskRequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		return nil, errors.New("body is not in correct format")
	}
	return &requestBody, nil
}

func validateNewTask(newTask newTaskRequestBody) error {
	hasRequiredPropertiesSet := checkRequiredProperties(newTask)
	if !hasRequiredPropertiesSet {
		return errors.New("task is not completely defined")
	}

	subtopicExists := checkIfSubtopicExists(newTask.SubtopicID)
	if !subtopicExists {
		return errors.New("subtopic does not exist")
	}

	return nil
}

func checkRequiredProperties(newTask newTaskRequestBody) bool {
	var hasName bool = false
	var hasOutput bool = false
	var hasExample bool = false
	hasName, newTask.Name = utils.GetTrimmedStringWithValue(newTask.Name)
	hasOutput, newTask.Output = utils.GetTrimmedStringWithValue(newTask.Output)
	hasExample, newTask.InputOutputExample = utils.GetTrimmedStringWithValue(newTask.InputOutputExample)

	var complexityNumber, err = strconv.Atoi(newTask.Complexity)
	if err != nil {
		return false
	}

	var hasTests bool = newTask.Tests != nil && len(newTask.Tests) > 0
	var hasComplexitySet bool = complexityNumber >= 1 && complexityNumber <= 5
	_, newTask.Input = utils.GetTrimmedStringWithValue(newTask.Input)
	_, newTask.SolutionCode = utils.GetTrimmedStringWithValue(newTask.SolutionCode)

	var hasRequiredPropertiesSet = hasName && hasOutput && hasExample && hasTests && hasComplexitySet
	return hasRequiredPropertiesSet
}

func checkIfSubtopicExists(subtopicID int) bool {
	var foundSubtopic = database.GetSubtopicById(subtopicID)
	return (foundSubtopic != nil)
}

func checkForValidTests(newTask newTaskRequestBody) error {
	if len(newTask.Tests) < 2 {
		return errors.New("too few tests")
	}

	if len(newTask.Tests) > 20 {
		return errors.New("too many tests")
	}

	for index, test := range newTask.Tests {
		var artefactIsExpected bool = false

		outputContainsChars, _ := utils.GetTrimmedStringWithValue(test.ExpectedOutput)
		artefactIsExpected, test.ExpectedOutput = utils.GetTrimmedStringWithValue(test.ArtefactSHA256)
		var outputDefined = outputContainsChars || artefactIsExpected

		if !outputDefined {
			return fmt.Errorf(fmt.Sprintf("test #%d is not testable", index+1))
		}
	}

	return nil
}

func checkForValidHelpSteps(newTask newTaskRequestBody) error {
	helpStepsCount := len(newTask.HelpSteps)

	if helpStepsCount < 1 {
		return errors.New("too few help steps")
	}

	if helpStepsCount > 10 {
		return errors.New("too many help steps")
	}

	for index, helpStep := range newTask.HelpSteps {
		containsCode, _ := utils.GetTrimmedStringWithValue(helpStep.HelperCode)
		containsText, _ := utils.GetTrimmedStringWithValue(helpStep.HelperText)

		if !(containsCode || containsText) {
			return fmt.Errorf(fmt.Sprintf("help step #%d is not well defined", index+1))
		}

		if !(helpStep.Step > 0 && helpStep.Step <= helpStepsCount) {
			return fmt.Errorf(fmt.Sprintf("help step #%d has a weird step number: %d", index+1, helpStep.Step))
		}
	}

	return nil
}

func attachCreatorIdToTask(r *http.Request, newTask *database.FullTask) (err error) {
	newTask.CreatorID, err = authentication.GetUserIdFromToken(r)
	return err
}

func storeTaskInDatabase(taskEntity *database.FullTask) error {
	err := database.SaveTask(taskEntity)
	return err
}

func attachTestsToTask(newTestDefinition []taskTestBody, taskEntity *database.FullTask) {
	for _, test := range newTestDefinition {
		var testEntity = database.TaskTest{
			Input:          test.Input,
			ExpectedOutput: test.ExpectedOutput,
			ArtefactSHA256: test.ArtefactSHA256,
		}

		taskEntity.Tests = append(taskEntity.Tests, testEntity)
	}
}

func attachHelpStepsToTask(taskHelpStepBody []taskHelpBody, taskEntity *database.FullTask) {
	for _, helpStep := range taskHelpStepBody {
		containsCode, trimmedHelperCode := utils.GetTrimmedStringWithValue(helpStep.HelperCode)
		containsText, trimmedHelperText := utils.GetTrimmedStringWithValue(helpStep.HelperText)

		var helpStepEntity = database.TaskHelpStep{
			Step:       helpStep.Step,
			HelperCode: sql.NullString{String: trimmedHelperCode, Valid: containsCode},
			HelperText: sql.NullString{String: trimmedHelperText, Valid: containsText},
		}

		taskEntity.HelpSteps = append(taskEntity.HelpSteps, helpStepEntity)
	}
}
