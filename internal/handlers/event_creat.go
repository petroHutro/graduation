package handlers

import (
	"encoding/json"
	"graduation/internal/entity"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"graduation/internal/utils"
	"net/http"
	"strconv"
	"time"

	"graduation/internal/encoding"
)

type Photo struct {
	Filename   string `json:"filename"`
	Base64Data []byte `json:"base64_data"`
}

type DataEventCreat struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Place        string  `json:"place"`
	Participants int     `json:"participants"`
	Date         string  `json:"date"`
	Active       bool    `json:"active"`
	Photo        []Photo `json:"photo"`
}

func HandlerEventCreat(w http.ResponseWriter, r *http.Request, st storage.Storage) {
	var data DataEventCreat

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.Error("bad json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.Header.Get("User_id"))
	if err != nil {
		logger.Error("cannot get user id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02 15:04", data.Date)
	if err != nil {
		logger.Error("cannot get data.Date: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := entity.Event{
		UserID:          userID,
		Title:           data.Title,
		Description:     data.Description,
		Place:           data.Place,
		Participants:    0,
		MaxParticipants: data.Participants,
		Date:            date,
		Active:          true,
	}

	for _, photo := range data.Photo {
		event.Images = append(event.Images, entity.Image{
			Filename:   utils.GenerateString() + ".jpg",
			Base64Data: photo.Base64Data,
		})
	}

	if err := st.CreateEvent(r.Context(), &event); err != nil {
		logger.Error("cannot creat event: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")

	w.Write([]byte(encoding.EncodeID(event.ID)))
}
