package taskexecution

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
)

var (
	ErrContainerTimeoutMark          = errors.New("timeout")
	ErrContainerForcefullyKilledMark = errors.New("forcefully killed")
	ErrIllegalOperation              = errors.New("attempted interaction with the system")
	ErrRunFailed                     = errors.New("run of compiled code failed")
	ErrFileSizeExceeded              = errors.New("file size limit exceeded")
)

const (
	STDIN_FILENAME_PREFIX     = "stdin_"
	STDOUT_FILENAME_PREFIX    = "stdout_"
	ARTEFACTS_FILENAME_PREFIX = "artefacts_"
	ERROR_FILENAME            = "error.txt"
)

type TestDataMismatchReason struct {
	TestInput      string `json:"testInput,omitempty"`
	Output         string `json:"output,omitempty"`
	ExpectedOutput string `json:"expectedOutput,omitempty"`
}

type TaskExecutionContainer struct {
	name             string
	executionData    *TaskExecutionData
	executionContext context.Context
	cancelFunc       *context.CancelFunc
}

func CreateTaskExecutionContainer(data *TaskExecutionData, timeoutContext context.Context) *TaskExecutionContainer {
	ctx, cancel := context.WithTimeout(timeoutContext, 30*time.Second)

	var container = &TaskExecutionContainer{
		name:             utils.CreateShortHash(data.tempFolderPath),
		executionData:    data,
		executionContext: ctx,
		cancelFunc:       &cancel,
	}

	return container
}

func (container *TaskExecutionContainer) runDockerRunnerImage() error {
	var volumeName string
	var tempDir string = container.executionData.tempFolderPath

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	volumeName = TASK_RUNNER_MOUNT_NAME
	sourceCodeFolderInContainer := filepath.Join(TASK_RUNNER_TEMP_TASKS_DIRECTORY, filepath.Base(tempDir))
	cpusForContainer := TASK_RUNNER_MAX_CPUS_AVAILABLE / float64(TASK_RUNNER_MAX_PARALLEL_AVAILABLE)

	var name = fmt.Sprintf("--name %s", container.name)
	var volumeAttachment = fmt.Sprintf("-v %s:%s", volumeName, TASK_RUNNER_TEMP_TASKS_DIRECTORY)
	var envFolder = fmt.Sprintf("-e SOURCE_CODE_FOLDER=%s", sourceCodeFolderInContainer)
	var envFile = fmt.Sprintf("-e SOURCE_FILE_NAME=%s", CPP_FILE_NAME)
	var envPrefixStdin = fmt.Sprintf("-e STDIN_FILENAME_PREFIX=%s", STDIN_FILENAME_PREFIX)
	var envPrefixStdout = fmt.Sprintf("-e STDOUT_FILENAME_PREFIX=%s", STDOUT_FILENAME_PREFIX)
	var envPrefixArtefacts = fmt.Sprintf("-e ARTEFACTS_FILENAME_PREFIX=%s", ARTEFACTS_FILENAME_PREFIX)
	var envErrorFilename = fmt.Sprintf("-e ERROR_FILENAME=%s", ERROR_FILENAME)
	var containerOptions = fmt.Sprintf("--memory %vm --cpus %v", TASK_RUNNER_MAX_RAM_MB_AVAILABLE, cpusForContainer)
	var securityOptions = "--rm --security-opt no-new-privileges --network none"

	var dockerRunArguments = strings.Join([]string{dockerPath,
		"run",
		name,
		volumeAttachment,
		envFolder,
		envFile,
		envPrefixStdin,
		envPrefixStdout,
		envPrefixArtefacts,
		envErrorFilename,
		containerOptions,
		securityOptions,
		"task-runner:latest",
	}, " ")

	var cmd = exec.Command("bash", "-c", dockerRunArguments)

	output, err := cmd.CombinedOutput()
	var stringOutput = string(output)
	if stringOutput != "" {
		if strings.Contains(stringOutput, "Unable to find image 'task-runner:latest'") {
			log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution"}).Error("Failed to build Docker container!")
		} else if strings.Contains(stringOutput, "Killed") {
			return ErrContainerForcefullyKilledMark
		} else {
			log.WithFields(log.Fields{"unexpected_output": stringOutput, "context": "task_execution"}).Warning("Unexpected output from task-runner container!")
		}
	}

	return err
}

func (container *TaskExecutionContainer) removeRunningDockerContainer() error {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf(dockerPath+" rm --force "+container.name))
	output, err := cmd.CombinedOutput()
	var stringOutputs = strings.Trim(string(output), "\n")
	if stringOutputs != container.name {
		if err == nil {
			return fmt.Errorf("container %s could not be stopped: %s", container.name, stringOutputs)
		} else {
			return fmt.Errorf("container %s could not be stopped: error:%v, output:%s", container.name, err, stringOutputs)
		}
	}

	return nil
}

func (container *TaskExecutionContainer) ReadError() string {
	contents, err := container.executionData.ReadErrorFile()
	if len(contents) == 0 || err != nil {
		return ""
	}
	return strings.TrimFunc(string(contents), func(r rune) bool { return unicode.IsSpace(r) })
}
