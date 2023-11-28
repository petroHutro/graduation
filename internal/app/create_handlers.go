package app

import (
	"graduation/internal/authorization"
	"graduation/internal/compression"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *App) createMiddlewareHandlers() {
	a.router.Use(logger.LoggingMiddleware)
	a.router.Use(compression.GzipMiddleware)
}

func (a *App) createHandlers() {
	a.router.Route("/api/event", func(r chi.Router) {
		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/creat", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventCreat(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventGet(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/dell/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventDell(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/close/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventClose(w, r, a.storage)
			})
	})

	a.router.Route("/api", func(r chi.Router) {
		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Get("/events", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventsGet(w, r, a.storage)
			})
	})

	a.router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerRegister(w, r, a.storage, a.conf.TokenSecretKey, a.conf.TokenEXP)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerLogin(w, r, a.storage, a.conf.TokenSecretKey, a.conf.TokenEXP)
		})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/add/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerUserAdd(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/dell/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerUserDell(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Get("/events", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerUserEvents(w, r, a.storage)
			})
	})

	a.router.Route("/api/images", func(r chi.Router) {
		r.Get("/{filename}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerImage(w, r, a.storage)
		})
	})
}
