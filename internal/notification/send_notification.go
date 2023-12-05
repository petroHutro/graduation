package notification

import (
	"context"
	"fmt"
	"graduation/internal/entity"
	"graduation/internal/mail"
	"sync"
	"time"
)

func (n *Notification) sendNotification(m *mail.Mail) error {
	date := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messages, err := n.storage.GetMessages(ctx, date.Add(3*time.Hour))
	if err != nil {
		return fmt.Errorf("cannot get message: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(messages))

	errCh := make(chan error, len(messages))

	for _, message := range messages {
		go func(message entity.Message) {
			for _, user := range message.Users {
				defer wg.Done()

				err := m.Send(user.Mail, message.Body, message.Urls)
				if err != nil {
					errCh <- fmt.Errorf("cannot send message: %w", err)
					return
				}

				err = n.storage.MessageUpdate(ctx, message.EventID, user.UserID)
				if err != nil {
					errCh <- fmt.Errorf("cannot update today: %w", err)
					return
				}
			}
		}(message)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
