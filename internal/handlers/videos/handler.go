package videos

import "github.com/go-chi/chi/v5"

func HandleVideos(r *chi.Mux) {
	r.Route("/videos", func(router chi.Router) {
		router.Get("/public", getPublicVideos)
		router.Get("/{videoId}", getSingleVideo)
	})
}
