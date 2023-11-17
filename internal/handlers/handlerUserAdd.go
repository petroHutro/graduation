package handlers

import (
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"strconv"
)

func HandlerUserAdd(w http.ResponseWriter, r *http.Request, st *storage.Storage) {
	// fmt.Println(r.URL.String()[14:])
	eventID, err := strconv.Atoi(r.URL.String()[14:])
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

	if err := st.AddEventUser(r.Context(), eventID, userID); err != nil {
		logger.Error("cannot add event user: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
