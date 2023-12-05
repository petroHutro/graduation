package handlerstest

import (
	"context"
	"errors"
	"graduation/internal/config"
	"graduation/internal/handlers"
	"graduation/internal/logger"
	"graduation/internal/storage/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerImage(t *testing.T) {
	type mockBehavior func(r *mock.MockStorage, ctx context.Context, filename string)

	if err := logger.InitLogger(config.Logger{LoggerFilePath: "file.log", LoggerFileFlag: false, LoggerMultiFlag: false}); err != nil {
		logger.Panic(err.Error())
	}

	tests := []struct {
		name                 string
		inputFilename        string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: `
POST /api/images #1 
correct inputFilename
got status 200
			`,
			inputFilename: "321",
			mockBehavior: func(r *mock.MockStorage, ctx context.Context, filename string) {
				r.EXPECT().GetImage(ctx, filename).Return("123", nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "123",
		},
		{
			name: `
POST /api/images #2
not correct return GetImage
got status 404
			`,

			mockBehavior: func(r *mock.MockStorage, ctx context.Context, filename string) {
				r.EXPECT().GetImage(ctx, filename).Return("", errors.New("err"))
			},
			expectedStatusCode: 404,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock.NewMockStorage(c)
			test.mockBehavior(repo, context.Background(), test.inputFilename)

			h := handlers.Init(repo, nil, "", 0)

			handler := func(w http.ResponseWriter, r *http.Request) {
				h.Image(w, r)
			}

			req, err := http.NewRequest("GET", "/api/images/"+test.inputFilename, nil)
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
