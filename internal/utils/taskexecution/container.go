package taskexecution

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
)

var (
	ErrContainerTimeoutMark          = errors.New("timeout")
	ErrContainerForcefullyKilledMark = errors.New("forcefully killed")
)

const CONTAINER_TIMEOUT_MARK = "timeout"
const CONTAINER_FORCEFULLY_KILLED_MARK = "forcefully killed"

type TaskExecutionContainer struct {
	executionData    *TaskExecutionData
	executionContext context.Context
	cancelFunc       *context.CancelFunc
}

func CreateTaskExecutionContainer(data *TaskExecutionData, timeoutContext context.Context) *TaskExecutionContainer {
	ctx, cancel := context.WithTimeout(timeoutContext, 30*time.Second)

	var container = &TaskExecutionContainer{
		executionData:    data,
		executionContext: ctx,
		cancelFunc:       &cancel,
	}

	return container
}

func (container *TaskExecutionContainer) RunTests() error {
	var errChan = make(chan error, 1)
	var sourceCodePath = container.executionData.File.Name()
	var containerName = utils.CreateShortHash(sourceCodePath)

	log.Debugf("Source code path: %s", sourceCodePath)

	go func() {
		errChan <- runDockerRunnerImage(containerName, sourceCodePath)
	}()

	select {
	case <-container.executionContext.Done():
		log.WithFields(log.Fields{"context": "task_execution"}).Warning("Timeout reached - forcefully removing Docker container!")

		err := removeRunningDockerContainer(containerName)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution"}).Error("Could not delete temp container after task execution timeout!")
		}

		return errors.New(CONTAINER_TIMEOUT_MARK)
	case err := <-errChan:
		return err
	}
}

func removeRunningDockerContainer(containerName string) error {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf(dockerPath+" rm --force "+containerName))
	output, err := cmd.CombinedOutput()
	var stringOutputs = strings.Trim(string(output), "\n")
	if stringOutputs != containerName {
		if err == nil {
			return fmt.Errorf("container %s could not be stopped: %s", containerName, stringOutputs)
		} else {
			return fmt.Errorf("container %s could not be stopped: error:%v, output:%s", containerName, err, stringOutputs)
		}
	}

	return nil
}

func runDockerRunnerImage(containerName string, fullFilePath string) error {
	volumeName := ""
	tempDirectory := filepath.Dir(fullFilePath)
	sourceFileName := filepath.Base(fullFilePath)
	parentFolderName := filepath.Base(tempDirectory)

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	if utils.IsProd() {
		volumeName = TASKS_VOLUME_NAME
	} else {
		volumeName = filepath.Dir(tempDirectory)
	}

	var options = "--rm --memory 50m --cpus 0.5 --security-opt no-new-privileges --network none task-runner:latest"
	var name = fmt.Sprintf("--name %s", containerName)
	var volumeAttachment = fmt.Sprintf("-v %s:%s", volumeName, TASK_RUNNER_TEMP_TASKS_DIRECTORY)
	var environmentVars = fmt.Sprintf("-e SOURCE_CODE_FOLDER=%s -e SOURCE_FILE_NAME=%s", parentFolderName, sourceFileName)
	var dockerRunArguments = strings.Join([]string{dockerPath, "run", name, volumeAttachment, environmentVars, options}, " ")

	var cmd = exec.Command("bash", "-c", dockerRunArguments)

	output, err := cmd.CombinedOutput()
	var stringOutput = string(output)
	if stringOutput != "" {
		if strings.Contains(stringOutput, "Unable to find image 'task-runner:latest'") {
			log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution"}).Error("Failed to build Docker container!")
		} else if strings.Contains(stringOutput, "Killed") {
			return errors.New(CONTAINER_FORCEFULLY_KILLED_MARK)
		} else {
			log.WithFields(log.Fields{"suspicious_output": stringOutput, "context": "task_execution"}).Warning("Suspicious output from task-runner container!")
		}
	}

	return err
}
