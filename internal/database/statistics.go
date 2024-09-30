package database

import (
	"database/sql"
)

type SolutionAttemptsDto struct {
	SolutionAttemptsPerSubtopic []SolutionAttemptsPerSubtopic `json:"solutionAttemptsPerSubtopic"`
}

type SolutionAttemptsPerSubtopic struct {
	Subtopic                        `json:"subtopic"`
	TotalTasksInSubtopicCount       int `json:"totalTasksInSubtopicCount"`
	SuccessfullyCompletedTasksCount int `json:"successfullyCompletedTasksCount"`
	TotalTriesPerThisSubtopic       int `json:"totalTries"`
}

func GetAllSolutionAttempts(userId int) (*SolutionAttemptsDto, error) {
	var taskExecutions []TaskExecution
	var aggregateObject SolutionAttemptsDto

	Instance.db.Where("id_user = @userId", sql.Named("userId", userId)).Preload("Task.Subtopic").Find(&taskExecutions)

	topics := GetAllTasksPerTopic()

	for _, topic := range topics {
		for _, subtopic := range topic.Subtopics {
			solutionAttempts := getExecutionStatsForSubtopic(taskExecutions, subtopic)
			aggregateObject.SolutionAttemptsPerSubtopic = append(aggregateObject.SolutionAttemptsPerSubtopic, solutionAttempts)
		}
	}

	return &aggregateObject, nil
}

func getExecutionStatsForSubtopic(taskExecutions []TaskExecution, subtopic *Subtopic) SolutionAttemptsPerSubtopic {
	var tasksInSubtopic = len(subtopic.Tasks)
	subtopic.Tasks = nil

	var totalTries = 0
	var completedTasks = make(map[int]bool)
	for _, taskExecution := range taskExecutions {
		if taskExecution.Task.SubtopicID == subtopic.ID {
			totalTries++
			if taskExecution.WasSuccessful {
				completedTasks[taskExecution.TaskID] = true
			}
		}
	}

	solutionAttempts := SolutionAttemptsPerSubtopic{
		Subtopic:                        *subtopic,
		TotalTasksInSubtopicCount:       tasksInSubtopic,
		SuccessfullyCompletedTasksCount: len(completedTasks),
		TotalTriesPerThisSubtopic:       totalTries,
	}
	return solutionAttempts
}
