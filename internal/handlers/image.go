package handlers

import (
	"graduation/internal/logger"
	"net/http"
)

// func errorResponse(w http.ResponseWriter, message string, err error, status int) {
// 	logger.Error("%s: %v", message, err)
// 	w.WriteHeader(status)
// }

// func goodResponse(w http.ResponseWriter, resp []byte, contentType string, status int) {
// 	w.WriteHeader(status)
// 	w.Header().Set("Content-Type", contentType)
// 	w.Write(resp)
// }

func (h *Handler) Image(w http.ResponseWriter, r *http.Request) {
	url, err := h.storage.GetImage(r.Context(), r.URL.String()[12:])
	if err != nil {
		logger.Error("photo not found: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(url))
}
