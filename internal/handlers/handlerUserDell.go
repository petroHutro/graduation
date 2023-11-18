package handlers

import (
	"errors"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"strconv"
)

func HandlerUserDell(w http.ResponseWriter, r *http.Request, st *storage.Storage) {
	eventID, err := strconv.Atoi(r.URL.String()[15:])
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

	if err := st.DellEventUser(r.Context(), eventID, userID); err != nil {
		var repErr *storage.RepError
		if errors.As(err, &repErr) {
			if repErr.UniqueViolation {
				logger.Error("user not add event: %v", err)
				w.WriteHeader(http.StatusConflict)
			} else if repErr.ForeignKeyViolation {
				logger.Error("event not exist: %v", err)
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			logger.Error("cannot dell user from event: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
