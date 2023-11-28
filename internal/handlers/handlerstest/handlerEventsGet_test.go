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
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerEventsGet(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, from, to time.Time, limit, page int)

	if err := logger.InitLogger(config.Logger{FilePath: "file.log", FileFlag: false, MultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tests := []struct {
		name                 string
		inputBody            string
		from                 time.Time
		to                   time.Time
		limit                int
		page                 int
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
			inputBody: `{
				"from": "2023-11-16",
				"to": "2023-11-30",
				"limit": 20,
				"page": 1
			}`,
			from:  utils.ParseDate("2023-11-16 00:00"),
			to:    time.Date(2023, 11, 30, 23, 59, 59, 0, time.UTC),
			limit: 20,
			page:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, from, to time.Time, limit, page int) {
				r.EXPECT().GetEvents(ctx, from, to, limit, page).Return(
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
							UserID:          2,
							Title:           "Title_2",
							Description:     "Description_2",
							Place:           "Place_2",
							Participants:    0,
							MaxParticipants: 1,
							Date:            utils.ParseDate("2023-11-28 00:01"),
							Active:          true,
							Images:          []entity.Image{},
						},
					}, 1, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"page":1,"pages":1,"events":[{"id":"XA==","title":"Title_1","description":"Description_1","place":"Place_1","participants":0,"max_participants":1,"data":"2023-11-28T00:01:00Z","active":true,"photo":null},{"id":"Xw==","title":"Title_2","description":"Description_2","place":"Place_2","participants":0,"max_participants":1,"data":"2023-11-28T00:01:00Z","active":true,"photo":null}]}`,
		},
		{
			name: `
GET /api/event #2
not correct inputID
got status 400
			`,
			inputBody:          ``,
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, from, to time.Time, limit, page int) {},
			expectedStatusCode: 400,
		},
		{
			name: `
GET /api/event #3
not correct return GetEvents
got status 404
			`,
			inputBody: `{
				"from": "2023-11-16",
				"to": "2023-11-30",
				"limit": 20,
				"page": 1
			}`,
			from:  utils.ParseDate("2023-11-16 00:00"),
			to:    time.Date(2023, 11, 30, 23, 59, 59, 0, time.UTC),
			limit: 20,
			page:  1,
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, from, to time.Time, limit, page int) {
				r.EXPECT().GetEvents(ctx, from, to, limit, page).Return(
					nil, 0, errors.New("err"))
			},
			expectedStatusCode: 404,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), test.from, test.to, test.limit, test.page)

			handler := func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlerEventsGet(w, r, repo)
			}

			req, err := http.NewRequest("GET", "/api/events", strings.NewReader(test.inputBody))
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
