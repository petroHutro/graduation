package handlers

import (
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
)

func HandlerImage(w http.ResponseWriter, r *http.Request, st storage.Storage) {
	url, err := st.GetImage(r.Context(), r.URL.String()[12:])
	if err != nil {
		logger.Error("photo not found: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(url))
}
