package handlers

import (
	"encoding/json"
	"graduation/internal/encoding"
	"graduation/internal/logger"
	"net/http"
	"strconv"
)

func (h *Handler) UserEvents(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("User_id"))
	if err != nil {
		logger.Error("cannot get user id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := h.storage.GetUserEvents(r.Context(), userID)
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

	respEvents, err := json.Marshal(dataResp)
	if err != nil {
		logger.Error("cannot json to byte: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	w.Write(respEvents)
}
