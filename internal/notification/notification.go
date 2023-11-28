package notification

import (
	"context"
	"graduation/internal/config"
	"graduation/internal/logger"
	"graduation/internal/mail"
	"graduation/internal/storage"
	"time"
)

func SendNotification(st storage.Storage, m *mail.Mail) error {
	date := time.Now()

	err := st.SendMessage(context.Background(), date.Add(3*time.Hour), m.Send)
	if err != nil {
		logger.Error("cannot send message:%v", err)
	}
	return nil
}

func LoopNotification(st storage.Storage, conf *config.SMTP) {
	tickerSend := time.NewTicker(1 * time.Hour)
	tickerGet := time.NewTicker(2 * time.Hour)
	defer tickerSend.Stop()
	defer tickerGet.Stop()

	var con *mail.Mail
	con, err := mail.Init(conf)
	if err != nil {
		logger.Error("cannot init mail: %v", err)
	}

	if err := st.EventsToday(context.Background(), time.Now()); err != nil {
		logger.Error("cannot get evens: %v", err)
	}

	if err := SendNotification(st, con); err != nil {
		logger.Error("cannot send message: %v", err)
	}

	logger.Info("Start Notification")
	for {
		select {
		case <-tickerGet.C:
			date := time.Now()
			if err := st.EventsToday(context.Background(), date.Add(6*time.Hour)); err != nil {
				logger.Error("cannot get evens: %v", err)
			}
		case <-tickerSend.C:
			if err := con.CheckConnection(); err != nil {
				con, err = mail.Init(conf)
				if err != nil {
					logger.Error("cannot send message: %v", err)
				}
				logger.Error("cannot: %v", err)
			} else {
				if err := SendNotification(st, con); err != nil {
					logger.Error("cannot send message: %v", err)
				}
			}
		}
	}
}
