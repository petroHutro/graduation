package authorization_test

import (
	"bytes"
	"graduation/internal/authorization"
	"graduation/internal/config"
	"graduation/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizationMiddleware(t *testing.T) {
	if err := logger.InitLogger(config.Logger{LoggerFilePath: "file.log", LoggerFileFlag: false, LoggerMultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	requestBody := `{"login":"test","password":"123"}`

	token, err := authorization.BuildJWTString("secretKey", time.Hour*3, 1)
	assert.NoError(t, err)

	tests := []struct {
		name               string
		cookie             http.Cookie
		expectedStatusCode int
	}{
		{
			name: `
AuthorizationMiddleware #1 
correct Cookie
got status 200
			`,
			cookie:             http.Cookie{Name: "Authorization", Value: token},
			expectedStatusCode: 200,
		},

		{
			name: `
AuthorizationMiddleware #2 
not correct Cookie
got status 401
			`,
			cookie:             http.Cookie{Name: "Authorization", Value: "bad_token"},
			expectedStatusCode: 401,
		},
		{
			name: `
AuthorizationMiddleware #3 
not Cookie
got status 401
			`,
			cookie:             http.Cookie{},
			expectedStatusCode: 401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBufferString(requestBody)
			request := httptest.NewRequest(http.MethodPost, "/", buf)
			request.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			request.AddCookie(&test.cookie)

			handler := authorization.AuthorizationMiddleware("secretKey")(mockHandler)

			handler.ServeHTTP(recorder, request)

			assert.Equal(t, test.expectedStatusCode, recorder.Code)
		})
	}
}
