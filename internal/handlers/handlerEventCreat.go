package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"graduation/internal/utils"
	"net/http"
	"strconv"
	"time"
)

type Photo struct {
	Filename   string `json:"filename"`
	Base64Data string `json:"base64_data"`
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

func HandlerEventCreat(w http.ResponseWriter, r *http.Request, st *storage.Storage) {
	var buf bytes.Buffer
	var data DataEventCreat

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logger.Error("not body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &data); err != nil {
		logger.Error("not byte to json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := saveImage(data.Photo); err != nil {
		logger.Error("cannot save photo: %v", err)
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

	event := storage.Event{
		User_id:         userID,
		Title:           data.Title,
		Description:     data.Description,
		Place:           data.Place,
		Participants:    0,
		MaxParticipants: data.Participants,
		Date:            date,
		Active:          true,
	}

	for _, photo := range data.Photo {
		event.Urls = append(event.Urls, photo.Filename)
	}

	if err := st.CreateEvent(r.Context(), &event); err != nil {
		logger.Error("cannot creat event: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Event:" + strconv.Itoa(event.ID)))
}

func saveImage(photos []Photo) error {
	for index, photo := range photos {
		filename := utils.GenerateString() + ".jpg"
		photos[index].Filename = filename
		err := utils.SaveBase64Image(filename, photo.Base64Data)
		if err != nil {
			return fmt.Errorf("cannot set database: %w", err)
		}
	}
	return nil
}
