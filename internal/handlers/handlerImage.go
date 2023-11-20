package handlers

import (
	"bytes"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"net/http"
	"time"
)

func HandlerImage(w http.ResponseWriter, r *http.Request, st *storage.Storage) {
	filename := r.URL.String()[12:]
	imageBytes, err := st.GetImage(r.Context(), filename)
	if err != nil {
		logger.Error("cannot get Image: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", "attachment; filename=file.jpg")

	http.ServeContent(w, r, "file.jpg", time.Now(), bytes.NewReader(imageBytes))
}
