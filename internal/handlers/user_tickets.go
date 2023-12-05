package handlers

import (
	"encoding/json"
	"graduation/internal/encoding"
	"graduation/internal/logger"
	"net/http"
	"strconv"
)

type RespTicket struct {
	Status  bool   `json:"status"`
	Token   string `json:"token"`
	EventID string `json:"eventID"`
}

func (h *Handler) UserTickets(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("User_id"))
	if err != nil {
		logger.Error("cannot get user id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tickets, err := h.storage.UserTickets(r.Context(), userID)
	if err != nil {
		logger.Error("cannot get tickets: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dataResp := []RespTicket{}
	for _, ticket := range tickets {
		dataResp = append(dataResp, RespTicket{
			Status:  ticket.Status,
			Token:   ticket.Token,
			EventID: encoding.EncodeID(ticket.EventID),
		})
	}

	respTickets, err := json.Marshal(dataResp)
	if err != nil {
		logger.Error("cannot json to byte: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	w.Write(respTickets)

}
