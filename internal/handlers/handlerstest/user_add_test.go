package handlerstest

import (
	"context"
	"errors"
	"graduation/internal/config"
	"graduation/internal/entity"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/storage"
	"graduation/internal/storage/mock"
	"graduation/internal/ticket"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerUserAdd(t *testing.T) {
	type mockBehaviorOne func(r *mock.MockStorage, ctx context.Context, tick *entity.Ticket)
	type mockBehaviorTwo func(r *mock.MockStorage, ctx context.Context, eventID int)

	if err := logger.InitLogger(config.Logger{LoggerFilePath: "file.log", LoggerFileFlag: false, LoggerMultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tick := ticket.Init(&config.TicketKey{TicketSecretKey: "123"})

	tests := []struct {
		name               string
		inputID            string
		headerID           string
		inputEventID       int
		inputUserID        int
		mockBehaviorOne    mockBehaviorOne
		mockBehaviorTwo    mockBehaviorTwo
		expectedStatusCode int
	}{
		{
			name: `
POST /api/user/add #1 
correct inputID, headerID
got status 200
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehaviorOne: func(r *mock.MockStorage, ctx context.Context, tick *entity.Ticket) {
				r.EXPECT().AddEventUser(ctx, gomock.Any()).Return(nil)
			},
			mockBehaviorTwo: func(r *mock.MockStorage, ctx context.Context, eventID int) {
				r.EXPECT().GetDateEvent(ctx, eventID).Return(1, nil)
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
			mockBehaviorOne:    func(r *mock.MockStorage, ctx context.Context, tick *entity.Ticket) {},
			mockBehaviorTwo:    func(r *mock.MockStorage, ctx context.Context, eventID int) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/user/add #3
not correct return AddEventUser
got status 400
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehaviorOne: func(r *mock.MockStorage, ctx context.Context, tick *entity.Ticket) {
				r.EXPECT().AddEventUser(ctx, gomock.Any()).Return(errors.New("err"))
			},
			mockBehaviorTwo: func(r *mock.MockStorage, ctx context.Context, eventID int) {
				r.EXPECT().GetDateEvent(ctx, eventID).Return(1, nil)
			},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/user/add #4
not correct return GetDateEvent (event not exist)
got status 404
			`,
			inputID:         `MQ==`,
			headerID:        "1",
			inputEventID:    1,
			inputUserID:     1,
			mockBehaviorOne: func(r *mock.MockStorage, ctx context.Context, tick *entity.Ticket) {},
			mockBehaviorTwo: func(r *mock.MockStorage, ctx context.Context, eventID int) {
				r.EXPECT().GetDateEvent(ctx, eventID).Return(0, &storage.RepError{Err: errors.New("err"), ForeignKeyViolation: true})
			},
			expectedStatusCode: 404,
		},
		{
			name: `
POST /api/user/add #5
not correct return AddEventUser (user already add event)
got status 409
			`,
			inputID:      `MQ==`,
			headerID:     "1",
			inputEventID: 1,
			inputUserID:  1,
			mockBehaviorOne: func(r *mock.MockStorage, ctx context.Context, tick *entity.Ticket) {
				r.EXPECT().AddEventUser(ctx, gomock.Any()).Return(&storage.RepError{Err: errors.New("err"), UniqueViolation: true})
			},
			mockBehaviorTwo: func(r *mock.MockStorage, ctx context.Context, eventID int) {
				r.EXPECT().GetDateEvent(ctx, eventID).Return(1, nil)
			},
			expectedStatusCode: 409,
		},
		{
			name: `
POST /api/user/add #6
not correct headerID
got status 400
			`,
			inputID:            `MQ==`,
			headerID:           "",
			mockBehaviorOne:    func(r *mock.MockStorage, ctx context.Context, tick *entity.Ticket) {},
			mockBehaviorTwo:    func(r *mock.MockStorage, ctx context.Context, eventID int) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/user/add #7
not correct return GetDateEvent
got status 400
			`,
			inputID:         `MQ==`,
			headerID:        "1",
			inputEventID:    1,
			inputUserID:     1,
			mockBehaviorOne: func(r *mock.MockStorage, ctx context.Context, tick *entity.Ticket) {},
			mockBehaviorTwo: func(r *mock.MockStorage, ctx context.Context, eventID int) {
				r.EXPECT().GetDateEvent(ctx, eventID).Return(0, errors.New("err"))
			},
			expectedStatusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)

			test.mockBehaviorTwo(repo, context.Background(), 1)
			test.mockBehaviorOne(repo, context.Background(), &entity.Ticket{})

			h := handlers.Init(repo, tick, "", 0)

			handler := func(w http.ResponseWriter, r *http.Request) {
				h.UserAdd(w, r)
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
