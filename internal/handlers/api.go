package handlers

import (
	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"

	health "github.com/mislavmatijevic/prog-demos-backend/internal/handlers/health"
	videos "github.com/mislavmatijevic/prog-demos-backend/internal/handlers/videos"
)

func Handler(r *chi.Mux) {
	setupCors(r)

	r.Use(chimiddle.StripSlashes)
	log.Debug("Setting up /videos handler...")
	videos.HandleVideos(r)
	log.Debug("Setting up /health handler...")
	health.HandleHealth(r)
}

func setupCors(r *chi.Mux) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://localhost:*", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
}
