package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

func Handler(r *chi.Mux) {
	r.Use(chimiddle.StripSlashes)

	r.Route("/", func(router chi.Router) {
		router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			var healthCheckResponse = struct {
				Status string
			}{
				Status: "Still rockin'!",
			}
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(healthCheckResponse)
		})
	})
}
