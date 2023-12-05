package app

import (
	"graduation/internal/authorization"
	"graduation/internal/compression"
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
				a.handler.EventCreat(w, r)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				a.handler.EventGet(w, r)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/dell/{id}", func(w http.ResponseWriter, r *http.Request) {
				a.handler.EventDell(w, r)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/close/{id}", func(w http.ResponseWriter, r *http.Request) {
				a.handler.EventClose(w, r)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Get("/valid/{id}", func(w http.ResponseWriter, r *http.Request) {
				a.handler.ValidTicket(w, r)
			})
	})

	a.router.Route("/api", func(r chi.Router) {
		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Get("/events", func(w http.ResponseWriter, r *http.Request) {
				a.handler.EventsGet(w, r)
			})
	})

	a.router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			a.handler.Register(w, r)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			a.handler.Login(w, r)
		})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/add/{id}", func(w http.ResponseWriter, r *http.Request) {
				a.handler.UserAdd(w, r)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Post("/dell/{id}", func(w http.ResponseWriter, r *http.Request) {
				a.handler.UserDell(w, r)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Get("/events", func(w http.ResponseWriter, r *http.Request) {
				a.handler.UserEvents(w, r)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.TokenSecretKey)).
			Get("/tickets", func(w http.ResponseWriter, r *http.Request) {
				a.handler.UserTickets(w, r)
			})
	})

	a.router.Route("/api/images", func(r chi.Router) {
		r.Get("/{filename}", func(w http.ResponseWriter, r *http.Request) {
			a.handler.Image(w, r)
		})
	})
}
