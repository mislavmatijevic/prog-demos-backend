package tasks

import (
	"github.com/go-chi/chi/v5"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
)

func HandleTasks(r *chi.Mux) {
	r.Route("/tasks", func(router chi.Router) {
		router.Get("/", getAllTasksPerTopics)
		router.Get("/{taskId}", getSingleTask)

		router.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authentication.RequireAccessToken)

			protectedRouter.Post("/{taskId}/run", executeTask)
		})

		router.Group(func(adminRouter chi.Router) {
			adminRouter.Use(authentication.RequireAccessToken)
			adminRouter.Use(authentication.RequireSpecialType)

			adminRouter.Post("/", createTask)
		})
	})
}
