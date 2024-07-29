package health

import (
	"github.com/go-chi/chi"
)

func HandleHealth(r *chi.Mux) {
	r.Route("/health", func(router chi.Router) {
		router.Get("/check", Check_Health)
	})
}
