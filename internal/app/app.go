package app

import (
	"fmt"
	"graduation/internal/authorization"
	"graduation/internal/config"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/notification"
	"graduation/internal/router"
	"graduation/internal/storage"

	"graduation/internal/compression"

	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type App struct {
	storage storage.Storage
	conf    *config.Flags
	router  *chi.Mux
}

func newApp() (*App, error) {
	conf := config.LoadServerConfigure()

	if err := logger.InitLogger(conf.Logger); err != nil {
		return nil, fmt.Errorf("cannot init logger: %w", err)
	}

	storage, err := storage.InitStorage(&conf.Storage)
	if err != nil {
		return nil, fmt.Errorf("cannot init storage: %w", err)
	}

	router := router.CreateRouter()
	logger.Info("Running server: address:%s port:%d", conf.Host, conf.Port)
	return &App{conf: conf, router: router, storage: storage}, nil
}

func (a *App) createMiddlewareHandlers() {
	a.router.Use(logger.LoggingMiddleware)
	a.router.Use(compression.GzipMiddleware)
}

func (a *App) createHandlers() {
	a.router.Route("/api/event", func(r chi.Router) {
		r.With(authorization.AuthorizationMiddleware(a.conf.SecretKey)).
			Post("/creat", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventCreat(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.SecretKey)).
			Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventGet(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.SecretKey)).
			Post("/dell/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventDell(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.SecretKey)).
			Post("/close/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventClose(w, r, a.storage)
			})
	})

	a.router.Route("/api", func(r chi.Router) {
		r.With(authorization.AuthorizationMiddleware(a.conf.SecretKey)).
			Get("/events", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventsGet(w, r, a.storage)
			})
	})

	a.router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerRegister(w, r, a.storage, a.conf.SecretKey, a.conf.TokenEXP)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerLogin(w, r, a.storage, a.conf.SecretKey, a.conf.TokenEXP)
		})

		r.With(authorization.AuthorizationMiddleware(a.conf.SecretKey)).
			Post("/add/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerUserAdd(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.SecretKey)).
			Post("/dell/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerUserDell(w, r, a.storage)
			})

		r.With(authorization.AuthorizationMiddleware(a.conf.SecretKey)).
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

func Run() error {
	app, err := newApp()
	if err != nil {
		return fmt.Errorf("cannot init app: %w", err)
	}
	defer logger.Shutdown()

	app.createMiddlewareHandlers()
	app.createHandlers()

	go notification.LoopNotification(app.storage)

	address := app.conf.Host + ":" + strconv.Itoa(app.conf.Port)

	return http.ListenAndServe(address, app.router)
}
