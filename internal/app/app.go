package app

import (
	"fmt"
	"graduation/internal/handlers"
	"graduation/internal/router"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type App struct {
	router *chi.Mux
}

func newApp() (*App, error) {
	router := router.CreateRouter()

	return &App{router: router}, nil
}

func (a *App) createMiddlewareHandlers() {
	// a.router.Use(logger.LoggingMiddleware)
}

func (a *App) createHandlers() {
	a.router.Route("/{id:[a-zA-Z0-9]+}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			handlers.Handler(w, r)
		})
	})
}

func Run() error {
	app, err := newApp()
	if err != nil {
		return fmt.Errorf("cannot init app: %w", err)
	}

	app.createMiddlewareHandlers()
	app.createHandlers()

	return http.ListenAndServe(":8080", app.router)
}
