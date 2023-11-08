package app

import (
	"fmt"
	"graduation/internal/config"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/router"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type App struct {
	conf   *config.Flags
	router *chi.Mux
}

func newApp() (*App, error) {
	conf := config.LoadServerConfigure()
	if err := logger.InitLogger(conf.Logger); err != nil {
		return nil, fmt.Errorf("cannot init logger: %w", err)
	}

	router := router.CreateRouter()

	return &App{conf: conf, router: router}, nil
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
	defer logger.Shutdown()

	app.createMiddlewareHandlers()
	app.createHandlers()

	address := app.conf.Host + ":" + strconv.Itoa(app.conf.Port)

	return http.ListenAndServe(address, app.router)
}
