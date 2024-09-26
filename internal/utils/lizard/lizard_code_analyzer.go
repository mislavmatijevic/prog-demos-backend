package lizard

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CodeScore = struct {
	Tokens     int     `json:"tokens"`
	Complexity int     `json:"complexity"`
	TotalScore float32 `json:"totalScore"`
}

/*
My system of calculating solution scores:

	score = (0.3 * totalCCN + 0.7 * totalTokens) / 2 + 10 * (numberOfFunctions - 1) - 15 * (isAvgTokens > 40).

I think many smaller functions are great and therefore I encourage usage of them in code.
*/
func CalculateScore(solutionCode *os.File) (*CodeScore, error) {
	lizardOutput, err := getLizardOutput(solutionCode.Name())
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

	var score = (0.3*totalCcn + 0.7*totalTokens) / 2
	log.Debugf("Original score: %v", score)
	score = awardManyFunctions(score, functionCount)
	score = punishHighAverageTokenCountPerFunction(score, averageTokensPerFunction)

	var roundedScore = float64(math.Round(score*100) / 100)

	return &CodeScore{Complexity: int(totalCcn), Tokens: int(totalTokens), TotalScore: float32(roundedScore)}, nil
}

func awardManyFunctions(score float64, functionCount float64) float64 {
	if functionCount > 1 {
		var awardMultiplier = functionCount
		if awardMultiplier > 10 {
			awardMultiplier = 10
		}
		var awardForManyFunctions = 10 * awardMultiplier
		score += awardForManyFunctions
	}
	return score
}

func punishHighAverageTokenCountPerFunction(score float64, averageTokensPerFunction float64) float64 {
	var punishmentForLongFunctions = float64(15 * int(averageTokensPerFunction/40))
	score -= punishmentForLongFunctions
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

	log.Errorf("Problematic lizard output: %s", lines)
	return 0, fmt.Errorf("unable to parse lizard's field %d", fieldIndex)
}
