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

func TestHandlerEventClose(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, eventID, userID int)

	if err := logger.InitLogger(config.Logger{FilePath: "file.log", FileFlag: false, MultiFlag: false}); err != nil {
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
POST /api/event/close #1 
correct inputID, headerID
got status 200
			`,
			inputID:      `XA==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().CloseEvent(ctx, eventID, userID).Return(nil)
			},
			expectedStatusCode: 200,
		},
		{
			name: `
POST /api/event/close #2
not correct inputID
got status 400
			`,
			inputID:            ``,
			headerID:           "1",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/event/close #3
not correct return CloseEvent
got status 400
			`,
			inputID:      `XA==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().CloseEvent(ctx, eventID, userID).Return(errors.New("err"))
			},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/event/close #4
not correct return setUser (event not exist)
got status 404
			`,
			inputID:      `XA==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().CloseEvent(ctx, eventID, userID).Return(&storage.RepError{Err: errors.New("err"), ForeignKeyViolation: true})
			},
			expectedStatusCode: 404,
		},
		{
			name: `
POST /api/event/close #5
not correct return setUser (user not have event)
got status 401
			`,
			inputID:      `XA==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().CloseEvent(ctx, eventID, userID).Return(&storage.RepError{Err: errors.New("err"), UniqueViolation: true})
			},
			expectedStatusCode: 401,
		},
		{
			name: `
POST /api/event/close #6
not correct headerID
got status 400
			`,
			inputID:            `XA==`,
			headerID:           "",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {},
			expectedStatusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), test.inputEventID, test.inputUserID)

			handler := func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventClose(w, r, repo)
			}

			req, err := http.NewRequest("POST", "/api/event/close/"+test.inputID, nil)
			assert.NoError(t, err)

			req.Header.Set("User_id", test.headerID)

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
		})
	}
}
