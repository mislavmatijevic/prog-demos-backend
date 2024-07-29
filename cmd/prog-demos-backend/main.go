package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)

	var port = "8000"
	fmt.Printf("I'm rockin' at port %s!", port)
	err := http.ListenAndServe(fmt.Sprintf("localhost:%s", port), r)
	if err != nil {
		log.Error(err)
	}
}
