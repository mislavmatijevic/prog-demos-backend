package handlers

import (
	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/auth"
	health "github.com/mislavmatijevic/prog-demos-backend/internal/handlers/health"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/statistics"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/tasks"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/topics"
	videos "github.com/mislavmatijevic/prog-demos-backend/internal/handlers/videos"
	"github.com/mislavmatijevic/prog-demos-backend/internal/logging"
	"github.com/mislavmatijevic/prog-demos-backend/internal/logging/logging_requests"
)

func Handler(r *chi.Mux) {
	setupCors(r)

	r.Use(chimiddle.StripSlashes)
	r.Use(authentication.AttachTokenToRequest)
	r.Use(logging.HandleSecureEndpoints)
	r.Use(logging_requests.LogRequest)
	log.Debug("Setting up /auth handler...")
	auth.HandleAuth(r)
	log.Debug("Setting up /videos handler...")
	videos.HandleVideos(r)
	log.Debug("Setting up /tasks handler...")
	tasks.HandleTasks(r)
	log.Debug("Setting up /topics handler...")
	topics.HandleTopics(r)
	log.Debug("Setting up /statistics handler...")
	statistics.HandleStatistics(r)
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
