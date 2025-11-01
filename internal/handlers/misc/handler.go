package misc

import "github.com/go-chi/chi/v5"

func HandleMiscRoutes(r *chi.Mux) {
	r.Route("/", func(router chi.Router) {
		router.Get("/news", getNews)
		router.Post("/report-issue", reportIssue)
	})
}
