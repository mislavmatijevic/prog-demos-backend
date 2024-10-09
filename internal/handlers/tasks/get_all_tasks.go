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
		fillBasicTasksWithPersonalizedInfo(r, topics)
	}

	var res = tasksResponse{
		Success: true,
		Topics:  topics,
	}

	api.RespondOk(w, res)
}

func fillBasicTasksWithPersonalizedInfo(r *http.Request, topics []database.Topic) {
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
				task.BestExecutionForUser = database.GetBestScoreExecutionForUserAndTask(userId, task.ID)
				if task.BestExecutionForUser != nil {
					task.BestExecutionForUser.SubmittedCode = ""
				}
			}
		}
	}
}
