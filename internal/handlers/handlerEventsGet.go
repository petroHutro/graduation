package handlers

import (
	"encoding/json"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"time"
)

// type RespEvent struct {
// 	ID           int       `json:"id"`
// 	Title        string    `json:"title"`
// 	Description  string    `json:"description"`
// 	Place        string    `json:"place"`
// 	Participants int       `json:"participants"`
// 	Date         time.Time `json:"data"`
// 	Active       bool      `json:"active"`
// 	Photo        []string  `json:"photo"`
// }

type OptionalFrom time.Time

func (o *OptionalFrom) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}

	*o = OptionalFrom(date)
	return nil
}

type OptionalTo time.Time

func (o *OptionalTo) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}

	*o = OptionalTo(date.Add(23*time.Hour + 59*time.Minute + 59*time.Second))
	return nil
}

type DataEventsGet struct {
	From  OptionalFrom `json:"from"`
	To    OptionalTo   `json:"to"`
	Limit int          `json:"limit"`
	Page  int          `json:"page"`
}

func intDataEventsGet() *DataEventsGet {
	data := DataEventsGet{
		From:  OptionalFrom(time.Now().Truncate(24 * time.Hour)),
		To:    OptionalTo(time.Now().Truncate(24 * time.Hour).Add(23*time.Hour + 59*time.Minute + 59*time.Second)),
		Limit: 100,
		Page:  1,
	}
	return &data
}

func HandlerEventsGet(w http.ResponseWriter, r *http.Request, st *storage.Storage) {
	data := intDataEventsGet()

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.Error("not byte to json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := st.GetEvents(r.Context(), time.Time(data.From), time.Time(data.To), data.Limit, data.Page)
	if err != nil {
		logger.Error("cannot get events: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dataResp := []RespEvent{}
	for _, event := range events {
		dataResp = append(dataResp, RespEvent{
			ID:              event.ID,
			Title:           event.Title,
			Description:     event.Description,
			Place:           event.Place,
			Participants:    event.Participants,
			MaxParticipants: event.MaxParticipants,
			Date:            event.Date,
			Active:          event.Active,
			Photo:           event.Urls,
		})
	}

	respEvent, err := json.Marshal(dataResp)
	if err != nil {
		logger.Error("cannot json to byte: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	w.Write(respEvent)
}
