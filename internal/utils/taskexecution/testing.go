package taskexecution

import (
	"errors"
	"strings"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	log "github.com/sirupsen/logrus"
)

func (container *TaskExecutionContainer) RunTests() error {
	var errChan = make(chan error, 1)
	var sourceCodePath = container.executionData.File.Name()

	log.Debugf("Source code path: %s", sourceCodePath)

	go func() {
		errChan <- container.runDockerRunnerImage()
	}()

	select {
	case <-container.executionContext.Done():
		log.WithFields(log.Fields{"context": "task_execution"}).Warning("Timeout reached - forcefully removing Docker container!")

		err := container.removeRunningDockerContainer()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution"}).Error("Could not delete temp container after task execution timeout!")
		}

		return errors.New(CONTAINER_TIMEOUT_MARK)
	case err := <-errChan:
		if err == nil {
			err = container.checkForErrors()
		}

		return err
	}
}

func (container *TaskExecutionContainer) checkForErrors() (detectedError error) {
	var generatedError = container.ReadError()

	if generatedError == "Failed to run commandSegmentation fault" {
		detectedError = ErrIllegalOperation
	} else if generatedError != "" {
		detectedError = ErrRunFailed
	}

	return
}

func (container *TaskExecutionContainer) CheckOutputs() (*TestDataMismatchReason, error) {
	var ranTests = container.executionData.Tests

	for _, test := range ranTests {
		outputFileContents, err := container.executionData.ReadOutputFile(test)
		if err != nil {
			return nil, err
		}

		var actualTestOutput = string(outputFileContents)
		var contentsAreSame = compareExpectedAndActualTestOutputs(test.ExpectedOutput, actualTestOutput)

		if !contentsAreSame {
			return &TestDataMismatchReason{
				TestInput:      test.Input,
				Output:         actualTestOutput,
				ExpectedOutput: test.ExpectedOutput,
			}, nil
		}
	}

	return nil, nil
}

func compareExpectedAndActualTestOutputs(expected, actual string) (theyMatch bool) {
	theyMatch = true
	var expectedOutputs = strings.Split(expected, "\n")
	var actualNoNewlines = strings.ReplaceAll(actual, "\n", "")

	for _, expected := range expectedOutputs {
		theyMatch = strings.Contains(actualNoNewlines, expected)
		if !theyMatch {
			return
		}
	}

	return
}

func (container *TaskExecutionContainer) CheckArtefacts() (*TestDataMismatchReason, error) {
	var ranTests = container.executionData.Tests

	for _, test := range ranTests {
		var hasArtefacts bool = test.ArtefactSHA256.Valid
		if hasArtefacts {
			hashMatches, err := checkHashMatch(test, container.executionData)
			if err != nil {
				return nil, err
			} else if !hashMatches {
				return &TestDataMismatchReason{TestInput: test.Input}, nil
			}
		}
	}

	return nil, nil
}

func checkHashMatch(test database.TaskTest, executionData *TaskExecutionData) (hashMatches bool, err error) {
	artefactFileContents, err := executionData.ReadSha256FromArtefactFile(test)
	if err != nil {
		return false, err
	}

	actualArtefactSha256 := strings.Trim(string(artefactFileContents), "\n")
	expectedArtefactSha256 := test.ArtefactSHA256.String

	if actualArtefactSha256 != expectedArtefactSha256 {
		return false, nil
	}

	return true, nil
}
