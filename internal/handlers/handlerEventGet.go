package handlers

import (
	"encoding/json"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"strconv"
	"time"
)

type RespEvent struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Place           string    `json:"place"`
	Participants    int       `json:"participants"`
	MaxParticipants int       `json:"max_participants"`
	Date            time.Time `json:"data"`
	Active          bool      `json:"active"`
	Photo           []string  `json:"photo"`
}

func HandlerEventGet(w http.ResponseWriter, r *http.Request, st *storage.Storage) {
	eventID := r.URL.String()[11:]

	id, err := strconv.Atoi(eventID)
	if err != nil {
		logger.Error("cannot get id from url: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := st.GetEvent(r.Context(), id)
	if err != nil {
		logger.Error("cannot get event: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dataResp := RespEvent{
		ID:              event.ID,
		Title:           event.Title,
		Description:     event.Description,
		Place:           event.Place,
		Participants:    event.Participants,
		MaxParticipants: event.MaxParticipants,
		Date:            event.Date,
		Active:          event.Active,
		Photo:           event.Urls,
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
