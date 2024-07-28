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

	fmt.Println("I'm rockin' at port 8000!")
	err := http.ListenAndServe("localhost:8000", r)
	if err != nil {
		log.Error(err)
	}
}
