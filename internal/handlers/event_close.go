package handlers

import (
	"errors"
	"graduation/internal/encoding"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"strconv"
)

func (h *Handler) EventClose(w http.ResponseWriter, r *http.Request) {
	eventID, err := encoding.DecodeID(r.URL.String()[17:])
	if err != nil {
		logger.Error("cannot get id from url: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.Header.Get("User_id"))
	if err != nil {
		logger.Error("cannot get user id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.storage.CloseEvent(r.Context(), userID, eventID); err != nil {
		var repErr *storage.RepError
		if errors.As(err, &repErr) {
			if repErr.UniqueViolation {
				logger.Error("user not have event: %v", err)
				w.WriteHeader(http.StatusUnauthorized)
			} else if repErr.ForeignKeyViolation {
				logger.Error("event not exist: %v", err)
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			logger.Error("cannot close event: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
