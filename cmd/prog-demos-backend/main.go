package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers"
	"github.com/mislavmatijevic/prog-demos-backend/internal/mailing"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

func main() {
	setupLogging()
	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)

	var err error

	database.Initialize()

	err = godotenv.Load()
	if err != nil {
		log.Fatalln("Couldn't load env file!!")
	}

	authentication.Initialize()
	mailing.Initialize()

	var port = os.Getenv("PORT")
	listeningAddress := fmt.Sprintf("0.0.0.0:%s", port)
	log.Infof("I'm rockin' at %s!", listeningAddress)
	err = http.ListenAndServe(listeningAddress, r)
	if err != nil {
		log.Error(err)
	}
}

func setupLogging() {
	log.SetReportCaller(true)
	var formatter log.Formatter = nil

	if utils.IsProd() {
		log.SetLevel(log.WarnLevel)
		formatter = &log.JSONFormatter{PrettyPrint: true, TimestampFormat: time.RFC3339}
	} else {
		log.SetLevel(log.DebugLevel)
		formatter = &log.TextFormatter{ForceColors: true, TimestampFormat: time.StampMilli}
	}

	log.SetFormatter(formatter)
}
