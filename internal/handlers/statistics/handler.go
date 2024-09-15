package statistics

import (
	"github.com/go-chi/chi"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
)

func HandleStatistics(r *chi.Mux) {
	r.Route("/statistics", func(router chi.Router) {

		router.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authentication.RequireAccessToken)

			protectedRouter.Get("/solution-attempts", GetTotalCountOfSolutionAttempts)
		})
	})
}
