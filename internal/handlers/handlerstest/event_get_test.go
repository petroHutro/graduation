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

func TestHandlerEventGet(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, eventID int)

	if err := logger.InitLogger(config.Logger{LoggerFilePath: "file.log", LoggerFileFlag: false, LoggerMultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tests := []struct {
		name                 string
		inputID              string
		inputEventID         int
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: `
POST /api/event #1 
correct inputID
got status 200
			`,
			inputID:      `MQ==`,
			inputEventID: 1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID int) {
				r.EXPECT().GetEvent(ctx, eventID).Return(&entity.Event{
					ID:              1,
					UserID:          1,
					Title:           "Title",
					Description:     "Description",
					Place:           "Place",
					Participants:    0,
					MaxParticipants: 1,
					Date:            utils.ParseDate("2023-11-28 00:01"),
					Active:          true,
					Images:          []entity.Image{},
				}, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":"MQ==","title":"Title","description":"Description","place":"Place","participants":0,"max_participants":1,"data":"2023-11-28T00:01:00Z","active":true,"photo":null}`,
		},
		{
			name: `
POST /api/event #2
not correct inputID
got status 400
			`,
			inputID:            ``,
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, eventID int) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/event #3
not correct return GetEvent
got status 404
			`,
			inputID:      `MQ==`,
			inputEventID: 1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID int) {
				r.EXPECT().GetEvent(ctx, eventID).Return(nil, errors.New("err"))
			},
			expectedStatusCode: 404,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), test.inputEventID)

			handler := func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventGet(w, r, repo)
			}

			req, err := http.NewRequest("GET", "/api/event/"+test.inputID, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
			if test.expectedResponseBody != "" {
				assert.Equal(t, test.expectedResponseBody, rr.Body.String())
			}
		})
	}
}
