package handlerstest

import (
	"context"
	"errors"
	"graduation/internal/config"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"graduation/internal/storage/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerEventDell(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, userID, eventID int)

	if err := logger.InitLogger(config.Logger{LoggerFilePath: "file.log", LoggerFileFlag: false, LoggerMultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tests := []struct {
		name               string
		inputID            string
		headerID           string
		inputEventID       int
		inputUserID        int
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name: `
POST /api/event/dell #1 
correct inputID, headerID
got status 200
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, userID, eventID int) {
				r.EXPECT().DellEvent(ctx, userID, eventID).Return(nil)
			},
			expectedStatusCode: 200,
		},
		{
			name: `
POST /api/event/dell #2
not correct inputID
got status 400
			`,
			inputID:            ``,
			headerID:           "1",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, userID, eventID int) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/event/dell #3
not correct return DellEvent
got status 400
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, userID, eventID int) {
				r.EXPECT().DellEvent(ctx, userID, eventID).Return(errors.New("err"))
			},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/event/dell #4
not correct return DellEvent (event not exist)
got status 404
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, userID, eventID int) {
				r.EXPECT().DellEvent(ctx, userID, eventID).Return(&storage.RepError{Err: errors.New("err"), ForeignKeyViolation: true})
			},
			expectedStatusCode: 404,
		},
		{
			name: `
POST /api/event/dell #5
not correct return DellEvent (user not have event)
got status 401
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, userID, eventID int) {
				r.EXPECT().DellEvent(ctx, userID, eventID).Return(&storage.RepError{Err: errors.New("err"), UniqueViolation: true})
			},
			expectedStatusCode: 401,
		},
		{
			name: `
POST /api/event/dell #6
not correct headerID
got status 400
			`,
			inputID:            `MQ==`,
			headerID:           "",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, userID, eventID int) {},
			expectedStatusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), test.inputUserID, test.inputEventID)

			h := handlers.Init(repo, nil, "", 0)

			handler := func(w http.ResponseWriter, r *http.Request) {
				h.EventDell(w, r)
			}

			req, err := http.NewRequest("POST", "/api/event/dell/"+test.inputID, nil)
			assert.NoError(t, err)

			req.Header.Set("User_id", test.headerID)

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
		})
	}
}
