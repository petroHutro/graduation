package app

import (
	"fmt"
	"graduation/internal/logger"
	"net/http"
	"strconv"
)

func Run() error {
	app, err := newApp()
	if err != nil {
		return fmt.Errorf("cannot init app: %w", err)
	}
	defer logger.Shutdown()

	app.createMiddlewareHandlers()
	app.createHandlers()

	go app.notification.LoopNotification()

	address := app.conf.Host + ":" + strconv.Itoa(app.conf.Port)

	return http.ListenAndServe(address, app.router)
}
