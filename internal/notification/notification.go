package notification

import (
	"context"
	"fmt"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"time"
)

func LoopNotification(st *storage.Storage) {
	tickerSend := time.NewTicker(1 * time.Hour)
	tickerGet := time.NewTicker(12 * time.Hour)
	defer tickerSend.Stop()
	defer tickerGet.Stop()
	err := st.EventsToday(context.Background(), time.Now())
	if err != nil {
		logger.Error("cannot get evens:%v", err)
	}

	for {
		select {
		case <-tickerGet.C:
			date := time.Now()
			fmt.Println("tickerGet.C:")
			err := st.EventsToday(context.Background(), date.Add(12*time.Hour))
			if err != nil {
				logger.Error("cannot get evens:%v", err)
			}
		case <-tickerSend.C:
			date := time.Now()
			fmt.Println("tickerSend.C:")
			err := st.SendMessage(context.Background(), date.Add(3*time.Hour))
			if err != nil {
				logger.Error("cannot get evens:%v", err)
			}
		}
	}
}
