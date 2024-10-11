package topics

import "github.com/go-chi/chi/v5"

func HandleTopics(r *chi.Mux) {
	r.Route("/topics", func(router chi.Router) {
		router.Get("/", getAllTopics)
	})
}
