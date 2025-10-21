package auth

import "github.com/go-chi/chi/v5"

func HandleAuth(r *chi.Mux) {
	r.Route("/auth", func(router chi.Router) {
		router.Post("/login", loginUser)
		router.Post("/register", registerUser)
		router.Post("/activate", activateUser)
		router.Post("/refresh", refreshAccess)
		router.Post("/logout", logoutUser)
		router.Post("/password/request-reset", requestPasswordReset)
		router.Post("/password/reset/verify", checkPasswordResetToken)
		router.Post("/password/reset", resetPassword)
	})
}
