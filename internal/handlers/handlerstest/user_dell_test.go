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

func TestHandlerUserDell(t *testing.T) {
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
POST /api/user/dell #1 
correct inputID, headerID
got status 200
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().DellEventUser(ctx, eventID, userID).Return(nil)
			},
			expectedStatusCode: 200,
		},
		{
			name: `
POST /api/user/dell #2
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
POST /api/user/dell #3
not correct return DellEventUser
got status 400
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().DellEventUser(ctx, eventID, userID).Return(errors.New("err"))
			},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/user/dell #4
not correct return DellEventUser (event not exist)
got status 404
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().DellEventUser(ctx, eventID, userID).Return(&storage.RepError{Err: errors.New("err"), ForeignKeyViolation: true})
			},
			expectedStatusCode: 404,
		},
		{
			name: `
POST /api/user/dell #5
not correct return DellEventUser (user not add event)
got status 409
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, eventID, userID int) {
				r.EXPECT().DellEventUser(ctx, eventID, userID).Return(&storage.RepError{Err: errors.New("err"), UniqueViolation: true})
			},
			expectedStatusCode: 409,
		},
		{
			name: `
POST /api/user/dell #6
not correct headerID
got status 400
			`,
			inputID:            `MQ==`,
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

			h := handlers.Init(repo, nil, "", 0)

			handler := func(w http.ResponseWriter, r *http.Request) {
				h.UserDell(w, r)
			}

			req, err := http.NewRequest("POST", "/api/user/dell/"+test.inputID, nil)
			assert.NoError(t, err)

			req.Header.Set("User_id", test.headerID)

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
		})
	}
}
