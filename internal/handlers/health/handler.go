package health

import "github.com/go-chi/chi/v5"

func HandleHealth(r *chi.Mux) {
	r.Route("/health", func(router chi.Router) {
		router.Get("/check", checkHealth)
	})
}
