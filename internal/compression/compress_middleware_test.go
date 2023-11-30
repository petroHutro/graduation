package compression_test

import (
	"bytes"
	"compress/gzip"
	"graduation/internal/compression"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipMiddleware(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	requestBody := `{"login":"test","password":"123"}`

	testOne := "TEST 1 send gzip"
	t.Run(testOne, func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		assert.NoError(t, err)
		err = zb.Close()
		assert.NoError(t, err)

		request := httptest.NewRequest(http.MethodPost, "/", buf)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Encoding", "gzip")
		request.RequestURI = ""
		recorder := httptest.NewRecorder()

		handler := compression.GzipMiddleware(mockHandler)
		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)

		b, err := io.ReadAll(recorder.Body)
		assert.NoError(t, err)
		assert.NotNil(t, string(b))
	})

	testTwo := "TEST 2 accept gzip"
	t.Run(testTwo, func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)

		request := httptest.NewRequest(http.MethodPost, "/", buf)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept-Encoding", "gzip")
		request.RequestURI = ""
		recorder := httptest.NewRecorder()

		handler := compression.GzipMiddleware(mockHandler)
		handler.ServeHTTP(recorder, request)

		zr, err := gzip.NewReader(recorder.Body)
		assert.NoError(t, err)

		b, err := io.ReadAll(zr)
		assert.NoError(t, err)
		assert.NotNil(t, string(b))
	})
}
