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

	var database database.Instance
	err := database.Initialize()
	if err != nil {
		log.Error(err)
	}

	var port = os.Getenv("PORT")
	fmt.Printf("I'm rockin' at port %s!", port)
	err = http.ListenAndServe(fmt.Sprintf("localhost:%s", port), r)
	if err != nil {
		log.Error(err)
	}

	database.Close()
}
