package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"graduation/internal/authorization"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"time"

	"net/http"
)

type DataRegister struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func setAuthorization(secretKey string, tokenEXP time.Duration, id int) *http.Cookie {
	token, _ := authorization.BuildJWTString(secretKey, tokenEXP, id)
	cookie := http.Cookie{Name: "Authorization", Value: token, Path: "/api"}
	return &cookie
}

func HandlerRegister(w http.ResponseWriter, r *http.Request, st *storage.Storage, secretKey string, tokenEXP time.Duration) {
	var buf bytes.Buffer
	var data DataRegister

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logger.Error("not body :%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &data); err != nil {
		logger.Error("not byte to json :%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := st.SetUser(r.Context(), data.Login, data.Password)
	if err != nil {
		var repErr *storage.RepError
		if errors.As(err, &repErr) && repErr.Repetition {
			logger.Error("user already db: %v", err)
			w.WriteHeader(http.StatusConflict)
		} else {
			logger.Error("cannot set user: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	http.SetCookie(w, setAuthorization(secretKey, tokenEXP, userID))

	w.WriteHeader(http.StatusOK)
}
