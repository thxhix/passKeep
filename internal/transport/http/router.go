package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/thxhix/passKeeper/internal/transport/http/handlers"
	"github.com/thxhix/passKeeper/internal/transport/http/middleware"
)

func NewRouter(handlers handlers.Handlers, jwtParser middleware.TokenParser) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.GzipMiddleware)

	router.Route("/", func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/auth", func(r chi.Router) {
				r.Post("/register", handlers.Register)
				r.Post("/login", handlers.Login)
				r.Post("/refresh", handlers.Refresh)
			})

			r.Route("/keychain", func(r chi.Router) {
				r.Use(middleware.Authorize(jwtParser, &handlers))

				r.Get("/", handlers.GetKeys)

				r.Get("/{uuid}", handlers.GetKey)
				r.Delete("/{uuid}", handlers.DeleteKey)

				r.Post("/credential", handlers.AddCredential)
				r.Post("/card", handlers.AddCard)
				r.Post("/text", handlers.AddText)
				r.Post("/file", handlers.AddFile)
			})
		})
	})

	return router
}
