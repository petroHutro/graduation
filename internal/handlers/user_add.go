package handlers

import (
	"errors"
	"graduation/internal/encoding"
	"graduation/internal/entity"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"strconv"
)

func (h *Handler) UserAdd(w http.ResponseWriter, r *http.Request) {
	eventID, err := encoding.DecodeID(r.URL.String()[14:])
	if err != nil {
		logger.Error("cannot get eventID from url: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.Header.Get("User_id"))
	if err != nil {
		logger.Error("cannot get user id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ticket := entity.Ticket{
		UserID:  userID,
		EventID: eventID,
		Exp:     10,
	}

	if err := h.tick.Generate(&ticket); err != nil {
		logger.Error("cannot creat ticket: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.storage.AddEventUser(r.Context(), &ticket); err != nil {
		var repErr *storage.RepError
		if errors.As(err, &repErr) {
			if repErr.UniqueViolation {
				logger.Error("user already add event: %v", err)
				w.WriteHeader(http.StatusConflict)
			} else if repErr.ForeignKeyViolation {
				logger.Error("event not exist: %v", err)
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			logger.Error("cannot add event user: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(ticket.Token))
}
