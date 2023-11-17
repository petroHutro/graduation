package handlers

import (
	"bytes"
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

func HandlerLogin(w http.ResponseWriter, r *http.Request, st *storage.Storage, secretKey string, tokenEXP time.Duration) {
	var buf bytes.Buffer
	var data DataLogin

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logger.Error("not body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &data); err != nil {
		logger.Error("not byte to json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := st.GetUser(r.Context(), data.Login, data.Password)
	if err != nil {
		logger.Error("bad login or password: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, setAuthorization(secretKey, tokenEXP, userID))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data.Login + " " + data.Password))
}
