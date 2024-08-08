package tasks

import (
	"github.com/go-chi/chi"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
)

func HandleTasks(r *chi.Mux) {
	r.Route("/tasks", func(router chi.Router) {
		router.Get("/", GetAllTasksPerTopics)
		router.Get("/{taskId}", GetSingleTask)

		router.Group(func(r chi.Router) {
			r.Use(authentication.UseAuthenticator())
			r.Use(authentication.UseVerifier())

			router.Post("/{taskId}/run", ExecuteTask)
		})
	})
}
