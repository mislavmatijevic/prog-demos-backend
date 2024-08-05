package videos

import "github.com/go-chi/chi"

func HandleVideos(r *chi.Mux) {
	r.Route("/videos", func(router chi.Router) {
		router.Get("/public", GetPublicVideos)
		router.Get("/{videoId}", GetSingleVideo)
	})
}
