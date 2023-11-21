package handlers

import (
	"bytes"
	"errors"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"time"
)

func HandlerImage(w http.ResponseWriter, r *http.Request, st *storage.Storage) {
	filename := r.URL.String()[12:]
	imageBytes, err := st.GetImage(r.Context(), filename)
	if err != nil {
		var repErr *storage.RepError
		if errors.As(err, &repErr) && repErr.UniqueViolation {
			logger.Error("photo not found: %v", err)
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.Error("cannot get Image: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", "attachment; filename=file.jpg")

	http.ServeContent(w, r, "file.jpg", time.Now(), bytes.NewReader(imageBytes))
}
