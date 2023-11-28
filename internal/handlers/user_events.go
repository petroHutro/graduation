package handlers

import (
	"encoding/json"
	"graduation/internal/encoding"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"strconv"
)

func HandlerUserEvents(w http.ResponseWriter, r *http.Request, st storage.Storage) {
	userID, err := strconv.Atoi(r.Header.Get("User_id"))
	if err != nil {
		logger.Error("cannot get user id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := st.GetUserEvents(r.Context(), userID)
	if err != nil {
		logger.Error("cannot get events: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dataResp := []RespEvent{}
	for index, event := range events {
		dataResp = append(dataResp, RespEvent{
			ID:              encoding.EncodeID(event.ID),
			Title:           event.Title,
			Description:     event.Description,
			Place:           event.Place,
			Participants:    event.Participants,
			MaxParticipants: event.MaxParticipants,
			Date:            event.Date,
			Active:          event.Active,
		})
		for _, image := range event.Images {
			dataResp[index].Photo = append(dataResp[index].Photo, image.Filename)
		}
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