package notification

import (
	"context"
	"graduation/internal/logger"
	"graduation/internal/mail"
	"time"
)

func (n *Notification) LoopNotification() {
	tickerSend := time.NewTicker(1 * time.Hour)
	tickerGet := time.NewTicker(2 * time.Hour)
	defer tickerSend.Stop()
	defer tickerGet.Stop()

	var con *mail.Mail
	con, err := mail.Init(n.conf)
	if err != nil {
		logger.Error("cannot init mail: %v", err)
	}

	if err := n.storage.EventsToday(context.Background(), time.Now()); err != nil {
		logger.Error("cannot get evens: %v", err)
	}

	if err := n.sendNotification(con); err != nil {
		logger.Error("cannot send message: %v", err)
	}

	logger.Info("Start Notification")
	for {
		select {
		case <-tickerGet.C:
			date := time.Now()
			if err := n.storage.EventsToday(context.Background(), date.Add(6*time.Hour)); err != nil {
				logger.Error("cannot get evens: %v", err)
			}
		case <-tickerSend.C:
			if err := con.CheckConnection(); err != nil {
				con, err = mail.Init(n.conf)
				if err != nil {
					logger.Error("cannot send message: %v", err)
				}
				logger.Error("cannot: %v", err)
			} else {
				if err := n.sendNotification(con); err != nil {
					logger.Error("cannot send message: %v", err)
				}
			}
		}
	}
}
