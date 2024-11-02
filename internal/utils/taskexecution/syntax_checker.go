package taskexecution

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

var gppErrorRegex = regexp.MustCompile(`^.*:(\d+):(\d+):\s+error:\s+(.*)$`)

type GppCompilerReportedSyntaxError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

func CheckSyntax(execution *TaskExecutionData) (reportedErrors []GppCompilerReportedSyntaxError, err error) {
	reportedErrors, err = findAllErrorsInSolutionCode(execution.File)
	if err != nil {
		log.WithError(err).
			WithFields(
				log.Fields{
					"priority":  "medium",
					"context":   "task_execution",
					"file_name": execution.File.Name(),
					"task_id":   execution.InitializedTaskExecution.TaskID,
				},
			).
			Error("Couldn't check code syntax!")
	}
	return
}

func findAllErrorsInSolutionCode(cppFile *os.File) ([]GppCompilerReportedSyntaxError, error) {
	cmd := exec.Command("g++", "-fsyntax-only", "-o /dev/null", cppFile.Name())
	output, err := cmd.CombinedOutput()

	var foundErrors []GppCompilerReportedSyntaxError

	if err != nil {
		foundErrors, err = extractGppSyntaxErrors(string(output))
	}

	return foundErrors, err
}

func extractGppSyntaxErrors(compilerOutput string) ([]GppCompilerReportedSyntaxError, error) {
	var scanner = bufio.NewScanner(strings.NewReader(compilerOutput))
	var detectedErrors []GppCompilerReportedSyntaxError

	for scanner.Scan() {
		line := scanner.Text()

		matches := gppErrorRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			lineNum, _ := strconv.Atoi(matches[1])
			columnNum, _ := strconv.Atoi(matches[2])

			var newError = GppCompilerReportedSyntaxError{
				Line:    lineNum,
				Column:  columnNum,
				Message: matches[3],
			}

			detectedErrors = append(detectedErrors, newError)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return detectedErrors, nil
}
