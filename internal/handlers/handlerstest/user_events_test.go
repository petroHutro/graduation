package handlerstest

import (
	"context"
	"errors"
	"graduation/internal/config"
	"graduation/internal/entity"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/storage/mock"
	"graduation/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerUserEvents(t *testing.T) {
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
GET /api/event #1 
correct inputID
got status 200
			`,
			headerID:    "1",
			inputUserID: 1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, userID int) {
				r.EXPECT().GetUserEvents(ctx, userID).Return(
					[]entity.Event{
						{
							ID:              1,
							UserID:          1,
							Title:           "Title_1",
							Description:     "Description_1",
							Place:           "Place_1",
							Participants:    0,
							MaxParticipants: 1,
							Date:            utils.ParseDate("2023-11-28 00:01"),
							Active:          true,
							Images:          []entity.Image{},
						},
						{
							ID:              2,
							UserID:          1,
							Title:           "Title_2",
							Description:     "Description_2",
							Place:           "Place_2",
							Participants:    0,
							MaxParticipants: 1,
							Date:            utils.ParseDate("2023-11-28 00:01"),
							Active:          true,
							Images:          []entity.Image{},
						},
					}, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `[{"id":"MQ==","title":"Title_1","description":"Description_1","place":"Place_1","participants":0,"max_participants":1,"data":"2023-11-28T00:01:00Z","active":true,"photo":null},{"id":"Mg==","title":"Title_2","description":"Description_2","place":"Place_2","participants":0,"max_participants":1,"data":"2023-11-28T00:01:00Z","active":true,"photo":null}]`,
		},
		{
			name: `
GET /api/event #2
not correct inputID
got status 400
			`,
			headerID:           "",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, userID int) {},
			expectedStatusCode: 400,
		},
		{
			name: `
GET /api/event #3
not correct return GetEvents
got status 400
			`,
			headerID:    "1",
			inputUserID: 1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, userID int) {
				r.EXPECT().GetUserEvents(ctx, userID).Return(
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

			handler := func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerUserEvents(w, r, repo)
			}

			req, err := http.NewRequest("GET", "/api/user/events", nil)
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
