package notification

import (
	"context"
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

func LoopNotification(st storage.Storage) {
	// con, err := mail.Init()
	// if err != nil {
	// 	logger.Error("cannot send message: %v", err)
	// }
	// if err := SendNotification(st, con); err != nil {
	// 	logger.Error("cannot Send Notification: %v", err)
	// }
	logger.Info("Gooood")

	// 	tickerSend := time.NewTicker(1 * time.Hour)
	// 	tickerGet := time.NewTicker(12 * time.Hour)
	// 	defer tickerSend.Stop()
	// 	defer tickerGet.Stop()
	// 	if err := st.EventsToday(context.Background(), time.Now()); err != nil {
	// 		logger.Error("cannot get evens: %v", err)
	// 	}

	// 	var con *mail.Mail
	// 	con, err := mail.Init()
	// 	if err != nil {
	// 		logger.Error("cannot send message: %v", err)
	// 	}

	//	for {
	//		select {
	//		case <-tickerGet.C:
	//			date := time.Now()
	//			fmt.Println("tickerGet.C:")
	//			if err := st.EventsToday(context.Background(), date.Add(12*time.Hour)); err != nil {
	//				logger.Error("cannot get evens: %v", err)
	//			}
	//		case <-tickerSend.C:
	//			fmt.Println("tickerSend.C:")
	//			if err := con.CheckConnection(); err != nil {
	//				con, err = mail.Init()
	//				if err != nil {
	//					logger.Error("cannot send message: %v", err)
	//				}
	//				logger.Error("cannot: %v", err)
	//			} else {
	//				if err := SendNotification(st, con); err != nil {
	//					logger.Error("cannot send message: %v", err)
	//				}
	//			}
	//		}
	//	}
}
