package handlerstest

import (
	"context"
	"errors"
	"graduation/internal/config"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/storage/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerLogin(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, login, password string)

	if err := logger.InitLogger(config.Logger{FilePath: "file.log", FileFlag: false, MultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tests := []struct {
		name      string
		inputBody string

		inputLogin    string
		inputPassword string

		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name: `
POST /api/user/login #1 
correct input body
got status 200
			`,
			inputBody:     `{"login": "user_1", "password": "password_1"}`,
			inputLogin:    "user_1",
			inputPassword: "password_1",
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, login, password string) {
				r.EXPECT().GetUser(ctx, login, password).Return(1, nil)
			},
			expectedStatusCode: 200,
		},
		{
			name: `
POST /api/user/login #2
not correct input body
got status 400
			`,
			inputBody:          ``,
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, login, password string) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/user/login #3
not correct return setUser
got status 401
			`,
			inputBody:     `{"login": "user_1", "password": "password_1"}`,
			inputLogin:    "user_1",
			inputPassword: "password_1",
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, login, password string) {
				r.EXPECT().GetUser(ctx, login, password).Return(0, errors.New("err"))
			},
			expectedStatusCode: 401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), test.inputLogin, test.inputPassword)

			handler := func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerLogin(w, r, repo, "your_secret_key", time.Hour)
			}

			req, err := http.NewRequest("POST", "/api/user/login", strings.NewReader(test.inputBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
			if rr.Code == 200 {
				if rr.Header().Get("Set-Cookie") == "" {
					t.Errorf("handler did not set expected cookie")
				}
			}
		})
	}
}
