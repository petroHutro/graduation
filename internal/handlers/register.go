package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"graduation/internal/authorization"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"time"

	"net/http"
)

type DataRegister struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Mail     string `json:"mail"`
}

func setAuthorization(secretKey string, tokenEXP time.Duration, id int) (*http.Cookie, error) {
	token, err := authorization.BuildJWTString(secretKey, tokenEXP, id)
	if err != nil {
		return nil, fmt.Errorf("cannot get token: %v", err)
	}
	cookie := http.Cookie{Name: "Authorization", Value: token, Path: "/api"}
	return &cookie, nil
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var data DataRegister

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.Error("bad json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.storage.SetUser(r.Context(), data.Login, data.Password, data.Mail)
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

	token, err := setAuthorization(h.tokenSecretKey, h.tokenEXP, userID)
	if err != nil {
		logger.Error("cannot get token: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.SetCookie(w, token)

	w.WriteHeader(http.StatusOK)
}
