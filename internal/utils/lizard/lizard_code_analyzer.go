package lizard

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
)

const complexityWeight = 13
const tokenWeight = 17
const mainScoreMultiplier = 50
const maxFunctionsAwardMultiplier = 1
const manyFunctionsAward = 5
const longFunctionPenalty = 5
const awardPerTaskComplexityPoint = 500

type CodeScore struct {
	Tokens     int     `json:"tokens"`
	Complexity int     `json:"complexity"`
	TotalScore float32 `json:"totalScore"`
}

func (comparedWith *CodeScore) HasBetterScoreThan(compareTo *CodeScore) bool {
	return comparedWith.TotalScore > compareTo.TotalScore
}

/*
My system of calculating solution scores promotes usage of many smaller functions.
I think it's great and therefore I encourage usage of them in code.
*/
func CalculateScore(fileWithCode *os.File, taskComplexity int) (*CodeScore, error) {
	err := sanitizeFileContents(fileWithCode)
	if err != nil {
		return nil, err
	}

	lizardOutput, err := getLizardOutput(fileWithCode.Name())
	if err != nil {
		return nil, err
	}

	functionCount, err := parseFunctionCountFromLizardOutput(lizardOutput)
	if err != nil {
		return nil, err
	}

	averageTokensPerFunction, err := parseTokenCountFromLizardOutput(lizardOutput)
	if err != nil {
		return nil, err
	}

	averageCcnPerFunction, err := parseComplexityFromLizardOutput(lizardOutput)
	if err != nil {
		return nil, err
	}

	var totalTokens = averageTokensPerFunction * functionCount
	var totalCcn = averageCcnPerFunction * functionCount

	var score = (complexityWeight/(totalCcn+1) + tokenWeight/(totalTokens+1)) * mainScoreMultiplier
	log.Debugf("Original score: %v", score)
	score = awardManyFunctions(score, functionCount)
	score = punishHighAverageTokenCountPerFunction(score, averageTokensPerFunction)
	score = awardForComplexity(score, taskComplexity)

	var roundedScore = utils.RoundNumberDownToTwoDecimals(score)

	return &CodeScore{Complexity: int(totalCcn), Tokens: int(totalTokens), TotalScore: roundedScore}, nil
}

func sanitizeFileContents(fileWithCode *os.File) error {
	contents, err := os.ReadFile(fileWithCode.Name())
	if err != nil {
		return err
	}

	var sanatizedContents = strings.ReplaceAll(string(contents), "GENERATED CODE", "")
	sanatizedContents = strings.ReplaceAll(string(sanatizedContents), "#lizard forgives", "")

	return os.WriteFile(fileWithCode.Name(), []byte(sanatizedContents), 0444)
}

func awardManyFunctions(score float64, functionCount float64) float64 {
	if functionCount > 1 {
		var awardMultiplier = functionCount
		if awardMultiplier > maxFunctionsAwardMultiplier {
			awardMultiplier = maxFunctionsAwardMultiplier
		}
		var awardForManyFunctions = manyFunctionsAward * awardMultiplier
		score += awardForManyFunctions
	}
	return score
}

func punishHighAverageTokenCountPerFunction(score float64, averageTokensPerFunction float64) float64 {
	var punishmentForLongFunctions = float64(longFunctionPenalty * int(averageTokensPerFunction/40))
	score -= punishmentForLongFunctions
	return score
}

func awardForComplexity(score float64, taskComplexity int) float64 {
	var awardMultiplier = taskComplexity - 1

	if awardMultiplier > 0 {
		var awardForTaskComplexity = awardPerTaskComplexityPoint * (awardMultiplier)
		score += float64(awardForTaskComplexity)
	}

	return score
}

func getLizardOutput(file string) (string, error) {
	log.Debugf("Lizard for file %s", file)
	lizardPath, _ := exec.LookPath("lizard")
	cmd := exec.Command(lizardPath, "-l cpp", "--ignore_warnings 999", file)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func parseComplexityFromLizardOutput(output string) (float64, error) {
	return getFieldValueFromLizardOutput(output, 1)
}

func parseTokenCountFromLizardOutput(output string) (float64, error) {
	return getFieldValueFromLizardOutput(output, 3)
}

func parseFunctionCountFromLizardOutput(output string) (float64, error) {
	return getFieldValueFromLizardOutput(output, 4)
}

func getFieldValueFromLizardOutput(output string, fieldIndex int) (float64, error) {
	lines := strings.Split(output, "\n")

	if len(lines) > 1 {
		fields := strings.Fields(lines[len(lines)-2])
		log.Debug(fields)
		if len(fields) == 8 {
			fieldValue, err := strconv.ParseFloat(fields[fieldIndex], 32)
			if err != nil {
				return 0, err
			}
			log.Debugf("value found: %v", fieldValue)
			return float64(fieldValue), nil
		} else {
			return 0, fmt.Errorf("lizard's last line has unexpected format: %s", fields)
		}
	}

	log.WithFields(log.Fields{"priority": "medium", "context": "task_execution", "problematic_lizard_output": output, "failed_at_field_index": fieldIndex}).Error("Lizard's output could not be parsed!")
	return 0, fmt.Errorf("unable to parse lizard's field %d", fieldIndex)
}
