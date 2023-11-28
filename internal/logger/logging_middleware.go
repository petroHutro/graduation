package logger

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("cannot write: %w", err)
	}
	r.responseData.size += size
	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)
		duration := time.Since(start)
		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", responseData.status),
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size),
		}
		l.logger.Info("Received request", fields...)
	})
}
