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
	"github.com/mislavmatijevic/prog-demos-backend/internal/mailing"
	"github.com/mislavmatijevic/prog-demos-backend/internal/monitoring/loki"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security/captcha"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/taskexecution"
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "main"}).Panic("Couldn't load env file!")
	}
	setupLogging()

	database.Initialize()
	captcha.Initialize()
	authentication.Initialize()
	mailing.Initialize()
	taskexecution.Initialize()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	var listeningAddress = fmt.Sprintf("0.0.0.0:%s", port)
	log.Infof("I'm rockin' at %s!", listeningAddress)

	var router *chi.Mux = chi.NewRouter()
	handlers.Handler(router)
	err = http.ListenAndServe(listeningAddress, router)
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
		log.SetReportCaller(true)
	}

	log.SetFormatter(formatter)
}
