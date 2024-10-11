package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers"
	"github.com/mislavmatijevic/prog-demos-backend/internal/logging/loki"
	"github.com/mislavmatijevic/prog-demos-backend/internal/mailing"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "main"}).Panic("Couldn't load env file!")
	}
	setupLogging()
	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)

	database.Initialize()

	authentication.Initialize()
	mailing.Initialize()

	port := os.Getenv("PORT")
	var listeningAddress = fmt.Sprintf("0.0.0.0:%s", port)
	log.Infof("I'm rockin' at %s!", listeningAddress)
	err = http.ListenAndServe(listeningAddress, r)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "main", "ip_address": listeningAddress}).Error("Failed while running.")
	}
}

func setupLogging() {
	var formatter log.Formatter = nil

	log.ErrorKey = "error_object"

	if utils.IsProd() {
		log.SetLevel(log.TraceLevel)
		formatter = &log.JSONFormatter{PrettyPrint: true, TimestampFormat: time.RFC3339}
		if utils.UseLoki() {
			loki.InitializeLoki()
		}
		log.SetReportCaller(true)
	} else {
		log.SetLevel(log.TraceLevel)
		formatter = &log.TextFormatter{ForceColors: true, TimestampFormat: time.StampMilli}
		log.SetReportCaller(false)
	}

	log.SetFormatter(formatter)
}
