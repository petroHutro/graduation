package compression

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"graduation/internal/logger"
	"graduation/internal/utils"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			contentType := r.Header.Get("Content-Type")
			if contentType == "application/json" || contentType == "text/html" {
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}
		}
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				logger.Error("GzipMiddleware not body:%v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			data, err := io.ReadAll(cr)
			if err != nil {
				logger.Error("GzipMiddleware cannot read body:%v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(data))
			if utils.IsJSON(data) {
				r.Header.Set("Content-Type", "application/json")
			} else if utils.IsText(data) {
				r.Header.Set("Content-Type", "text/plain")
			} else {
				logger.Error("GzipMiddleware not correct content type:%v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer cr.Close()
		}
		next.ServeHTTP(ow, r)
	})
}
