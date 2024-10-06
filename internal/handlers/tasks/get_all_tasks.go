package tasks

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	log "github.com/sirupsen/logrus"
)

type tasksResponse struct {
	Success bool             `json:"success"`
	Topics  []database.Topic `json:"topics"`
}

func getAllTasksPerTopics(w http.ResponseWriter, r *http.Request) {
	topics := database.GetAllTasksPerTopic()

	if err := authentication.ValidateJwtTokenFromRequest(r); err == nil {
		handleAuthenticatedUserRequest(r, topics)
	}

	var res = tasksResponse{
		Success: true,
		Topics:  topics,
	}

	api.RespondOk(w, res)
}

func handleAuthenticatedUserRequest(r *http.Request, topics []database.Topic) {
	var userId, err = authentication.GetUserIdFromRequest(r)
	if err != nil {
		log.Errorf("Token validated, but couldn't extract user id: %v", err)
		return
	}

	fillInfoOnCompletedTasks(userId, topics)
}

func fillInfoOnCompletedTasks(userId int, topics []database.Topic) {
	for _, topic := range topics {
		for _, subtopic := range topic.Subtopics {
			for _, task := range subtopic.Tasks {
				task.TaskExecution = database.GetBestScoreExecutionOfUserForTask(userId, task.ID)
				if task.TaskExecution != nil {
					task.TaskExecution.SubmittedCode = ""
				}
			}
		}
	}
}
