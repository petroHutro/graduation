package handlers

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	link := r.URL.String()[1:]
	if link == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(link))
}
