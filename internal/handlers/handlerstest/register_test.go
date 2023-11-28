package handlerstest

import (
	"context"
	"errors"
	"graduation/internal/config"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"graduation/internal/storage/mock"
	"strings"
	"time"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerRegister(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, login, password, mail string)

	if err := logger.InitLogger(config.Logger{FilePath: "file.log", FileFlag: false, MultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tests := []struct {
		name      string
		inputBody string

		inputLogin    string
		inputPassword string
		inputmail     string

		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: `
POST /api/user/register #1 
correct input body
got status 200
			`,
			inputBody:     `{"login": "user_1", "password": "password_1", "mail": "mail_1@mail.ru"}`,
			inputLogin:    "user_1",
			inputPassword: "password_1",
			inputmail:     "mail_1@mail.ru",
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, login, password, mail string) {
				r.EXPECT().SetUser(ctx, login, password, mail).Return(1, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":1}`,
		},
		{
			name: `
POST /api/user/register #2
not correct input body
got status 400
			`,
			inputBody:          ``,
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, login, password, mail string) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/user/register #3
not correct return setUser
got status 400
			`,
			inputBody:     `{"login": "user_1", "password": "password_1", "mail": "mail_1@mail.ru"}`,
			inputLogin:    "user_1",
			inputPassword: "password_1",
			inputmail:     "mail_1@mail.ru",
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, login, password, mail string) {
				r.EXPECT().SetUser(ctx, login, password, mail).Return(0, errors.New("err"))
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"id":1}`,
		},
		{
			name: `
POST /api/user/register #4
not correct return setUser (user already db)
got status 409
			`,
			inputBody:     `{"login": "user_1", "password": "password_1", "mail": "mail_1@mail.ru"}`,
			inputLogin:    "user_1",
			inputPassword: "password_1",
			inputmail:     "mail_1@mail.ru",
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, login, password, mail string) {
				r.EXPECT().SetUser(ctx, login, password, mail).Return(0, &storage.RepError{Err: errors.New("err"), Repetition: true})
			},
			expectedStatusCode:   409,
			expectedResponseBody: `{"id":1}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), test.inputLogin, test.inputPassword, test.inputmail)

			handler := func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerRegister(w, r, repo, "your_secret_key", time.Hour)
			}

			req, err := http.NewRequest("POST", "/api/user/register", strings.NewReader(test.inputBody))
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
