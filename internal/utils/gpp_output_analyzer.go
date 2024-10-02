package utils

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
)

var gppErrorRegex = regexp.MustCompile(`^.*:(\d+):(\d+):\s+error:\s+(.*)$`)

type GppCompilerReportedSyntaxError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

func ExtractGppSyntaxErrors(compilerOutput string) ([]GppCompilerReportedSyntaxError, error) {
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
