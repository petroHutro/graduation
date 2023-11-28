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

func TestHandlerUserAdd(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, eventID, userID int)

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
POST /api/user/add #1 
correct inputID, headerID
got status 200
			`,
			inputID:      `XA==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().AddEventUser(ctx, eventID, userID).Return(nil)
			},
			expectedStatusCode: 200,
		},
		{
			name: `
POST /api/user/add #2
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
POST /api/user/add #3
not correct return AddEventUser
got status 400
			`,
			inputID:      `XA==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().AddEventUser(ctx, eventID, userID).Return(errors.New("err"))
			},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/user/add #4
not correct return AddEventUser (event not exist)
got status 404
			`,
			inputID:      `XA==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().AddEventUser(ctx, eventID, userID).Return(&storage.RepError{Err: errors.New("err"), ForeignKeyViolation: true})
			},
			expectedStatusCode: 404,
		},
		{
			name: `
POST /api/user/add #5
not correct return AddEventUser (user already add event)
got status 409
			`,
			inputID:      `XA==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().AddEventUser(ctx, eventID, userID).Return(&storage.RepError{Err: errors.New("err"), UniqueViolation: true})
			},
			expectedStatusCode: 409,
		},
		{
			name: `
POST /api/user/add #6
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
				handlers.HandlerUserAdd(w, r, repo)
			}

			req, err := http.NewRequest("POST", "/api/user/add/"+test.inputID, nil)
			assert.NoError(t, err)

			req.Header.Set("User_id", test.headerID)

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
		})
	}
}