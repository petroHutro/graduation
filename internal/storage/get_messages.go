package storage

import (
	"context"
	"fmt"
	"graduation/internal/entity"
	"graduation/internal/utils"
	"time"
)

func (s *storageData) getMail(ctx context.Context, userID int) (string, error) {
	var mail string
	err := s.db.QueryRowContext(ctx, `
		SELECT mail 
		FROM users WHERE id = $1;
	`, userID).Scan(&mail)

	if err != nil {
		return "", fmt.Errorf("cannot scan: %w", err)
	}

	return mail, nil
}

func (s *storageData) getBody(ctx context.Context, event *entity.Event) (string, error) {
	body, err := utils.GenerateHTML(event)
	if err != nil {
		return "", fmt.Errorf("cannot get event: %w", err)
	}

	return body, nil
}

func (s *storageData) getUserToday(ctx context.Context, date time.Time) ([]entity.Message, error) {
	formattedTime := date.Format("2006-01-02 15:04:05")

	rowsEvent, err := s.db.QueryContext(ctx, `
		SELECT event_id, user_id
		FROM today
		WHERE send = FALSE 
		AND ABS(EXTRACT(EPOCH FROM ($1 - date)) / 3600) <= 3
		ORDER BY date;
	`, formattedTime)
	if err != nil || rowsEvent.Err() != nil {
		return nil, fmt.Errorf("cannot get user_id: %w", err)
	}
	defer rowsEvent.Close()

	uniqueEvent := make(map[int]bool)
	var messages []entity.Message
	for rowsEvent.Next() {
		var userID, eventID int
		err := rowsEvent.Scan(&eventID, &userID)
		if err != nil {
			return nil, fmt.Errorf("cannot scan: %w", err)
		}

		if _, exists := uniqueEvent[eventID]; !exists {
			message := entity.Message{
				EventID: eventID,
			}
			uniqueEvent[eventID] = true
			messages = append(messages, message)
		}

		for i := range messages {
			if messages[i].EventID == eventID {
				messageTo := entity.MessageTo{
					UserID: userID,
				}
				messages[i].Users = append(messages[i].Users, messageTo)
			}
		}
	}

	return messages, nil
}

func (s *storageData) GetMessages(ctx context.Context, date time.Time) ([]entity.Message, error) {
	messages, err := s.getUserToday(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("cannot get user today: %w", err)
	}

	for index, message := range messages {
		event, err := s.GetEvent(ctx, message.EventID)
		if err != nil {
			return nil, fmt.Errorf("cannot get event: %w", err)
		}

		var urls []string
		for index, image := range event.Images {
			url, err := s.GetImage(ctx, image.Filename)
			if err != nil {
				return nil, fmt.Errorf("cannot get url: %w", err)
			}
			event.Images[index].Filename = url
			urls = append(urls, url)
		}

		body, err := s.getBody(ctx, event)
		if err != nil {
			return nil, fmt.Errorf("cannot get body: %w", err)
		}

		for index, user := range message.Users {
			mail, err := s.getMail(ctx, user.UserID)
			if err != nil {
				return nil, fmt.Errorf("cannot get mail: %w", err)
			}
			message.Users[index].Mail = mail
		}

		messages[index].Body = body
		messages[index].Urls = urls
	}

	return messages, nil
}
