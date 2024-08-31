package tasks

import (
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

type newTestDefinition struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expectedOutput"`
	ArtefactSHA256 string `json:"artefactSha256,omitempty"`
}

type newTaskRequestBody struct {
	SubtopicID         int                 `json:"idSubtopic"`
	Name               string              `json:"name"`
	Complexity         string              `json:"complexity"`
	Input              string              `json:"input"`
	Output             string              `json:"output"`
	InputOutputExample string              `json:"inputOutputExample"`
	IsFinalBoss        bool                `json:"isFinalBoss"`
	StarterCode        string              `json:"starterCode"`
	Step1Code          string              `json:"step1Code,omitempty"`
	Step2Code          string              `json:"step2Code,omitempty"`
	Step3Code          string              `json:"step3Code,omitempty"`
	Helper1Text        string              `json:"helper1Text,omitempty"`
	Helper2Text        string              `json:"helper2Text,omitempty"`
	Helper3Text        string              `json:"helper3Text,omitempty"`
	SolutionCode       string              `json:"solutionCode,omitempty"`
	Tests              []newTestDefinition `json:"tests"`
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
		StarterCode:        body.StarterCode,
		Step1Code:          body.Step1Code,
		Step2Code:          body.Step2Code,
		Step3Code:          body.Step3Code,
		Helper1Text:        body.Helper1Text,
		Helper2Text:        body.Helper2Text,
		Helper3Text:        body.Helper3Text,
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
	if err != nil || requestBody.SolutionCode == "" {
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
	var hasStarterCode bool = false
	hasName, newTask.Name = utils.GetTrimmedStringWithValue(newTask.Name)
	hasOutput, newTask.Output = utils.GetTrimmedStringWithValue(newTask.Output)
	hasExample, newTask.InputOutputExample = utils.GetTrimmedStringWithValue(newTask.InputOutputExample)
	hasStarterCode, newTask.StarterCode = utils.GetTrimmedStringWithValue(newTask.StarterCode)

	var complexityNumber, err = strconv.Atoi(newTask.Complexity)
	if err != nil {
		return false
	}

	var hasTests bool = newTask.Tests != nil && len(newTask.Tests) > 0
	var hasComplexitySet bool = complexityNumber >= 1 && complexityNumber <= 5
	_, newTask.Input = utils.GetTrimmedStringWithValue(newTask.Input)
	_, newTask.Step1Code = utils.GetTrimmedStringWithValue(newTask.Step1Code)
	_, newTask.Step2Code = utils.GetTrimmedStringWithValue(newTask.Step2Code)
	_, newTask.Step3Code = utils.GetTrimmedStringWithValue(newTask.Step3Code)
	_, newTask.Helper1Text = utils.GetTrimmedStringWithValue(newTask.Helper1Text)
	_, newTask.Helper2Text = utils.GetTrimmedStringWithValue(newTask.Helper2Text)
	_, newTask.Helper3Text = utils.GetTrimmedStringWithValue(newTask.Helper3Text)
	_, newTask.SolutionCode = utils.GetTrimmedStringWithValue(newTask.SolutionCode)

	var hasRequiredPropertiesSet = hasName && hasOutput && hasExample && hasStarterCode && hasTests && hasComplexitySet
	return hasRequiredPropertiesSet
}

func checkIfSubtopicExists(subtopicID int) bool {
	var foundSubtopic = database.GetSubtopicById(subtopicID)
	return (foundSubtopic != nil)
}

func checkForValidTests(newTask newTaskRequestBody) error {
	if len(newTask.Tests) > 20 {
		return errors.New("too many tests")
	}

	for index, test := range newTask.Tests {
		var artefactIsExpected bool = false

		outputContainsChars, _ := utils.GetTrimmedStringWithValue(test.ExpectedOutput)
		artefactIsExpected, test.ExpectedOutput = utils.GetTrimmedStringWithValue(test.ArtefactSHA256)
		var outputDefined = outputContainsChars || artefactIsExpected

		if !outputDefined {
			return fmt.Errorf(fmt.Sprintf("test #%d is not testable", index))
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

func attachTestsToTask(newTestDefinition []newTestDefinition, taskEntity *database.FullTask) {
	for _, test := range newTestDefinition {
		var testEntity = database.Test{
			Input:          test.Input,
			ExpectedOutput: test.ExpectedOutput,
			ArtefactSHA256: test.ArtefactSHA256,
		}

		taskEntity.Tests = append(taskEntity.Tests, testEntity)
	}
}
