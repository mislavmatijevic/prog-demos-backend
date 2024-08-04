package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers"
)

func main() {
	log.SetReportCaller(true)
	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)

	database.Initialize()

	var port = os.Getenv("PORT")
	listeningAddress := fmt.Sprintf("0.0.0.0:%s", port)
	log.Infof("I'm rockin' at %s!", listeningAddress)
	err := http.ListenAndServe(listeningAddress, r)
	if err != nil {
		log.Error(err)
	}
}
