package app

import (
	"fmt"
	"graduation/internal/config"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/notification"
	"graduation/internal/router"
	"graduation/internal/storage"
	"graduation/internal/ticket"

	"github.com/go-chi/chi/v5"
)

type App struct {
	storage      storage.Storage
	conf         *config.Flags
	router       *chi.Mux
	tick         *ticket.TicketToken
	handler      *handlers.Handler
	notification *notification.Notification
}

func newApp() (*App, error) {
	conf, err := config.LoadServerConfigure()
	if err != nil {
		return nil, fmt.Errorf("cannot load server configure: %w", err)
	}

	if err := logger.InitLogger(conf.Logger); err != nil {
		return nil, fmt.Errorf("cannot init logger: %w", err)
	}

	storage, err := storage.InitStorage(&conf.Storage)
	if err != nil {
		return nil, fmt.Errorf("cannot init storage: %w", err)
	}

	tick := ticket.Init(&conf.TicketKey)

	router := router.CreateRouter()

	handler := handlers.Init(storage, tick, conf.TokenSecretKey, conf.TokenEXP)

	notification := notification.Init(storage, &conf.SMTP)

	logger.Info("Running server: address:%s port:%d", conf.Host, conf.Port)

	return &App{
		conf:         conf,
		router:       router,
		storage:      storage,
		tick:         tick,
		handler:      handler,
		notification: notification,
	}, nil
}
