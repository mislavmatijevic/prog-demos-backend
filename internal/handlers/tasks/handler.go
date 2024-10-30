package tasks

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimiddle "github.com/go-chi/chi/v5/middleware"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
)

func HandleTasks(r *chi.Mux) {
	r.Route("/tasks", func(router chi.Router) {
		router.Get("/", getAllTasksPerTopics)
		router.Get("/{taskIdentifier}", getSingleTask)
		router.Get("/{taskId}/help/{helpStep}", getHelpStep)

		router.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authentication.RequireAccessToken)
			protectedRouter.Use(chimiddle.ThrottleBacklog(10, 20, 30*time.Second))

			protectedRouter.Post("/{taskId}/run", executeTask)
		})

		router.Group(func(adminRouter chi.Router) {
			adminRouter.Use(authentication.RequireAccessToken)
			adminRouter.Use(authentication.RequireSpecialType)

			adminRouter.Post("/", createTask)
		})
	})
}
