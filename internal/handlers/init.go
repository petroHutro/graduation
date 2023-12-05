package handlers

import (
	"graduation/internal/storage"
	"graduation/internal/ticket"
	"time"
)

type Handler struct {
	storage        storage.Storage
	tick           *ticket.TicketToken
	tokenSecretKey string
	tokenEXP       time.Duration
}

func Init(storage storage.Storage, tick *ticket.TicketToken, tokenSecretKey string, tokenEXP time.Duration) *Handler {
	return &Handler{
		storage:        storage,
		tick:           tick,
		tokenSecretKey: tokenSecretKey,
		tokenEXP:       tokenEXP,
	}
}
