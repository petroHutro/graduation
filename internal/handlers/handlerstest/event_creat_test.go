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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerEventCreat(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, e *entity.Event)

	if err := logger.InitLogger(config.Logger{LoggerFilePath: "file.log", LoggerFileFlag: false, LoggerMultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tests := []struct {
		name                 string
		inputBody            string
		headerID             string
		event                entity.Event
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: `
POST /api/event/creat #1 
correct inputBody
got status 200
			`,
			inputBody: `{
				"title": "Title",
				"description": "Description",
				"place": "Place",
				"participants": 1,
				"date": "2023-11-28 00:00",
				"photo": []
				}`,
			headerID: "1",
			event: entity.Event{
				UserID:          1,
				Title:           "Title",
				Description:     "Description",
				Place:           "Place",
				Participants:    0,
				MaxParticipants: 1,
				Date:            utils.ParseDate("2023-11-28 00:00"),
				Active:          true,
			},
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, e *entity.Event) {
				r.EXPECT().CreateEvent(ctx, e).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "MA==",
		},
		{
			name: `
POST /api/event/creat #2
not correct inputBody
got status 400
					`,
			inputBody:          ``,
			headerID:           "1",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, e *entity.Event) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/event/creat #3
not correct headerID
got status 400
					`,
			inputBody: `{
				"title": "Title",
				"description": "Description",
				"place": "Place",
				"participants": 1,
				"date": "2023-11-28 00:00",
				"photo": []
				}`,
			headerID:           "",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, e *entity.Event) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/event/close #4 
nor correct date
got status 400
			`,
			inputBody: `{
				"title": "Title",
				"description": "Description",
				"place": "Place",
				"participants": 1,
				"date": "00:00",
				"photo": []
				}`,
			headerID:           "1",
			mockBehavior:       func(r *mock.MockStorage, ctx context.Context, e *entity.Event) {},
			expectedStatusCode: 400,
		},
		{
			name: `
POST /api/event/creat #5 
not correct return CreateEvent
got status 400
			`,
			inputBody: `{
				"title": "Title",
				"description": "Description",
				"place": "Place",
				"participants": 1,
				"date": "2023-11-28 00:00",
				"photo": []
				}`,
			headerID: "1",
			event: entity.Event{
				UserID:          1,
				Title:           "Title",
				Description:     "Description",
				Place:           "Place",
				Participants:    0,
				MaxParticipants: 1,
				Date:            utils.ParseDate("2023-11-28 00:00"),
				Active:          true,
			},
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, e *entity.Event) {
				r.EXPECT().CreateEvent(ctx, e).Return(errors.New("err"))
			},
			expectedStatusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), &test.event)

			h := handlers.Init(repo, nil, "", 0)

			handler := func(w http.ResponseWriter, r *http.Request) {
				h.EventCreat(w, r)
			}

			req, err := http.NewRequest("POST", "/api/event/crat", strings.NewReader(test.inputBody))
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
