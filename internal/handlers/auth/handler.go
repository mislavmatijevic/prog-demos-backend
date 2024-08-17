package auth

import "github.com/go-chi/chi"

func HandleAuth(r *chi.Mux) {
	r.Route("/auth", func(router chi.Router) {
		router.Post("/login", LoginUser)
		router.Post("/register", RegisterUser)
		router.Post("/activate", ActivateUser)
		router.Post("/refresh", RefreshAccess)
		router.Post("/logout", LogoutUser)
		router.Post("/password/request-reset", RequestPasswordReset)
	})
}
