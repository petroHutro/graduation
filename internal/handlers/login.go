package handlers

import (
	"encoding/json"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"time"

	"net/http"
)

type DataLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func HandlerLogin(w http.ResponseWriter, r *http.Request, st storage.Storage, secretKey string, tokenEXP time.Duration) {
	var data DataLogin

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.Error("bad json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := st.GetUser(r.Context(), data.Login, data.Password)
	if err != nil {
		logger.Error("bad login or password: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := setAuthorization(secretKey, tokenEXP, userID)
	if err != nil {
		logger.Error("cannot get token: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.SetCookie(w, token)

	w.WriteHeader(http.StatusOK)
}
