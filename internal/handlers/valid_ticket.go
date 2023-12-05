package handlers

import (
	"encoding/json"
	"errors"
	"graduation/internal/encoding"
	"graduation/internal/entity"
	"graduation/internal/logger"
	"net/http"
	"time"
)

type RespValid struct {
	Status bool      `json:"status"`
	ID     string    `json:"id"`
	Title  string    `json:"title"`
	Place  string    `json:"place"`
	Date   time.Time `json:"data"`
	Active bool      `json:"active"`
}

func (h *Handler) ValidTicket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.String()[17:]
	if token == "" {
		logger.Error("token from url emty: %v", errors.New("token emty"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ticket := entity.Ticket{
		Token: token,
	}

	if err := h.tick.Validate(&ticket); err != nil {
		logger.Error("cannot validate ticket: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := h.storage.GetEvent(r.Context(), ticket.EventID)
	if err != nil {
		logger.Error("cannot event: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dataResp := RespValid{
		Status: ticket.Status,
		ID:     encoding.EncodeID(event.ID),
		Title:  event.Title,
		Place:  event.Place,
		Date:   event.Date,
		Active: event.Active,
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
