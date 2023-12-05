package handlerstest

import (
	"context"
	"errors"
	"graduation/internal/config"
	"graduation/internal/entity"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/storage/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_UserTickets(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, userID int)

	if err := logger.InitLogger(config.Logger{LoggerFilePath: "file.log", LoggerFileFlag: false, LoggerMultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tests := []struct {
		name                 string
		headerID             string
		inputUserID          int
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: `
GET /api/tickets #1 
correct inputID
got status 200
			`,
			headerID:    "1",
			inputUserID: 1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, userID int) {
				r.EXPECT().UserTickets(ctx, userID).Return(
					[]entity.Ticket{
						{
							UserID:  1,
							EventID: 1,
							Exp:     1,
							Status:  true,
							Token:   "token_1",
						},
						{
							UserID:  1,
							EventID: 2,
							Exp:     1,
							Status:  false,
							Token:   "token_2",
						},
					}, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `[{"status":true,"token":"token_1","eventID":"MQ=="},{"status":false,"token":"token_2","eventID":"Mg=="}]`,
		},
		{
			name: `
GET /api/tickets #2
not correct inputID
got status 400
			`,
			headerID:           "",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, userID int) {},
			expectedStatusCode: 400,
		},
		{
			name: `
GET /api/tickets #3
not correct return GetEvents
got status 400
			`,
			headerID:    "1",
			inputUserID: 1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, userID int) {
				r.EXPECT().UserTickets(ctx, userID).Return(
					nil, errors.New("err"))
			},
			expectedStatusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), test.inputUserID)

			h := handlers.Init(repo, nil, "", 0)

			handler := func(w http.ResponseWriter, r *http.Request) {
				h.UserTickets(w, r)
			}

			req, err := http.NewRequest("GET", "/api/user/tickets", nil)
			assert.NoError(t, err)

			req.Header.Set("User_id", test.headerID)

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
			if test.expectedResponseBody != "" {
				assert.Equal(t, test.expectedResponseBody, rr.Body.String())
			}
		})
	}
}
