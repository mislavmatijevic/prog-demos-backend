package tasks

import "github.com/go-chi/chi"

func HandleTasks(r *chi.Mux) {
	r.Route("/tasks", func(router chi.Router) {
		router.Get("/", GetAllTasksPerTopics)
		router.Get("/{taskId}", GetSingleTask)
	})
}
