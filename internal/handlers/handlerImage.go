package handlers

import (
	"net/http"
)

func HandlerImage(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.String()[12:]
	http.ServeFile(w, r, "/Users/petro/GoProjects/graduation/images/"+filename)
}
