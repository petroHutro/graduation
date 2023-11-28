package app

import (
	"fmt"
	"graduation/internal/config"
	"graduation/internal/logger"
	"graduation/internal/router"
	"graduation/internal/storage"

	"github.com/go-chi/chi/v5"
)

type App struct {
	storage storage.Storage
	conf    *config.Flags
	router  *chi.Mux
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

	router := router.CreateRouter()
	logger.Info("Running server: address:%s port:%d", conf.Host, conf.Port)
	return &App{conf: conf, router: router, storage: storage}, nil
}
