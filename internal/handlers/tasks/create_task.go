package tasks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
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
	Identifier         int            `json:"identifier"`
	SubtopicID         int            `json:"idSubtopic"`
	Name               string         `json:"name"`
	Complexity         string         `json:"complexity"`
	Input              string         `json:"input"`
	Output             string         `json:"output"`
	InputOutputExample string         `json:"inputOutputExample"`
	IsBossBattle       bool           `json:"isBossBattle"`
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
		BasicTask: database.BasicTask{
			ID:           0,
			Identifier:   body.Identifier,
			SubtopicID:   body.SubtopicID,
			Name:         body.Name,
			Complexity:   body.Complexity,
			IsBossBattle: body.IsBossBattle,
		},
		Input:              body.Input,
		Output:             body.Output,
		InputOutputExample: body.InputOutputExample,
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

	err = validateNewTask(newTask)
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
		log.WithError(err).
			WithFields(log.Fields{"priority": "medium",
				"context":                "tasks",
				"problematic_task":       fmt.Sprintf("%v", taskEntity),
				"problematic_creator_id": taskEntity.CreatorID}).
			Error("Failed while attaching creator ID to task.")
		return
	}

	err = attachTestsToTask(newTask.Tests, taskEntity)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	if !newTask.IsBossBattle {
		err = checkForValidHelpSteps(*newTask)
		if err != nil {
			api.RequestErrorHandlerCustomMsg(w, err.Error())
			return
		}
	}

	err = storeTaskInDatabase(taskEntity)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		log.WithError(err).
			WithFields(log.Fields{
				"priority":         "medium",
				"context":          "tasks",
				"problematic_task": fmt.Sprintf("%v", taskEntity),
			}).
			Error("Failed while persisting task entity in database.")
		return
	}

	if !newTask.IsBossBattle {
		err = attachHelpStepsToTask(newTask.HelpSteps, taskEntity)
		if err != nil {
			api.InternalErrorHandlerCustomMsg(w, "Created task, but couldn't attach help steps!")
			log.WithError(err).
				WithFields(log.Fields{"priority": "medium",
					"context":                "tasks",
					"problematic_task":       fmt.Sprintf("%v", taskEntity),
					"problematic_help_steps": newTask.HelpSteps}).
				Error("Failed while attaching help steps to a created task.")
			return
		}
	}

	var res = responseBody{
		Success: true,
		Message: fmt.Sprintf(
			"Created new task (Identifier: %v, ID %v)",
			taskEntity.BasicTask.Identifier,
			taskEntity.BasicTask.ID,
		),
	}

	api.RespondOk(w, res)
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

func validateNewTask(newTask *newTaskRequestBody) error {
	if newTask.Identifier < 0 {
		return errors.New("desired identifier invalid")
	}

	if newTask.Identifier == 0 {
		largestIdentifierYet, err := database.GetLargestStoredTaskIdentifier()
		if err != nil {
			return err
		}
		newlyAssignedIdentifier := largestIdentifierYet + 1
		newTask.Identifier = newlyAssignedIdentifier
	}

	identifierUsed := checkIfIdentifierUsed(newTask.Identifier)
	if identifierUsed {
		return errors.New("desired identifier already in use")
	}

	hasRequiredPropertiesSet := checkRequiredProperties(*newTask)
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

	var hasRequiredPropertiesSet = hasName && hasOutput && hasExample && hasTests && hasComplexitySet
	return hasRequiredPropertiesSet
}

func checkIfSubtopicExists(subtopicID int) bool {
	var foundSubtopic = database.GetSubtopicById(subtopicID)
	return (foundSubtopic != nil)
}

func checkIfIdentifierUsed(desiredIdentifier int) bool {
	var foundTaskByIdentifier = database.GetSingleFullTaskByIdentifier(desiredIdentifier)
	return foundTaskByIdentifier != nil
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

	sort.SliceStable(newTask.HelpSteps, func(i, j int) bool {
		return newTask.HelpSteps[i].Step < newTask.HelpSteps[j].Step
	})

	for index, helpStep := range newTask.HelpSteps {
		containsCode, _ := utils.GetTrimmedStringWithValue(helpStep.HelperCode)
		containsText, _ := utils.GetTrimmedStringWithValue(helpStep.HelperText)

		if index+1 != helpStep.Step {
			return fmt.Errorf(fmt.Sprintf("help step #%d does not have an expected index %d", helpStep.Step, index+1))
		}

		if !(containsCode || containsText) {
			return fmt.Errorf(fmt.Sprintf("help step #%d is not well defined", helpStep.Step))
		}

		if !(helpStep.Step > 0 && helpStep.Step <= helpStepsCount) {
			return fmt.Errorf(fmt.Sprintf("help step #%d has a weird step number: %d", helpStep.Step, helpStep.Step))
		}
	}

	return nil
}

func attachCreatorIdToTask(r *http.Request, newTask *database.FullTask) (err error) {
	newTask.CreatorID, err = authentication.GetUserIdFromRequest(r)
	return err
}

func storeTaskInDatabase(taskEntity *database.FullTask) error {
	err := database.SaveTask(taskEntity)
	return err
}

func attachTestsToTask(newTestDefinition []taskTestBody, taskEntity *database.FullTask) error {
	for _, test := range newTestDefinition {
		containsSha256, artefactsHash := utils.GetTrimmedStringWithValue(test.ArtefactSHA256)
		var isValidSha256 = containsSha256 && len(artefactsHash) == 64

		if !isValidSha256 && len(artefactsHash) > 0 {
			return errors.New("expected artefact hash is not sha256")
		}

		var testEntity = database.TaskTest{
			Input:          test.Input,
			ExpectedOutput: test.ExpectedOutput,
			ArtefactSHA256: database.WrappedNullString{NullString: sql.NullString{String: artefactsHash, Valid: isValidSha256}},
		}

		taskEntity.Tests = append(taskEntity.Tests, testEntity)
	}

	return nil
}

func attachHelpStepsToTask(taskHelpStepBody []taskHelpBody, taskEntity *database.FullTask) error {
	for _, helpStep := range taskHelpStepBody {
		containsCode, trimmedHelperCode := utils.GetTrimmedStringWithValue(helpStep.HelperCode)
		containsText, trimmedHelperText := utils.GetTrimmedStringWithValue(helpStep.HelperText)

		var helpStepEntity = database.TaskHelpStep{
			IDTask:     taskEntity.ID,
			Step:       helpStep.Step,
			HelperCode: database.WrappedNullString{NullString: sql.NullString{String: trimmedHelperCode, Valid: containsCode}},
			HelperText: database.WrappedNullString{NullString: sql.NullString{String: trimmedHelperText, Valid: containsText}},
		}

		err := database.SaveHelpStep(&helpStepEntity)
		if err != nil {
			return err
		}
	}

	return nil
}
