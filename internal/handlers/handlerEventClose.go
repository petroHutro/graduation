package handlers

import (
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"strconv"
)

func HandlerEventClose(w http.ResponseWriter, r *http.Request, st *storage.Storage) {
	eventID, err := strconv.Atoi(r.URL.String()[17:])
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

	if err := st.CloseEvent(r.Context(), userID, eventID); err != nil { // проверить на 401 и 404
		logger.Error("cannot get event: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
