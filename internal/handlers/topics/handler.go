package topics

import (
	"github.com/go-chi/chi"
)

func HandleTopics(r *chi.Mux) {
	r.Route("/topics", func(router chi.Router) {
		router.Get("/", getAllTopics)
	})
}
