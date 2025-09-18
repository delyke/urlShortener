package handler

import (
	"bytes"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/mocks"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_HandleAPIShortenBatch(t *testing.T) {
	type args struct {
		body        []byte
		contentType string
		method      string
		request     string
	}

	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Batch Post Empty Body",
			args: args{
				body:        nil,
				contentType: "application/json",
				method:      http.MethodPost,
				request:     "/api/shorten/batch",
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name: "Batch Post Invalid JSON",
			args: args{
				body:        []byte(`{"invalid":"json"}`),
				contentType: "application/json",
				method:      http.MethodPost,
				request:     "/api/shorten/batch",
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name: "Batch Post Success",
			args: args{
				body: []byte(`[
					{"correlation_id":"1", "original_url":"https://yandex.com"},
					{"correlation_id":"2", "original_url":"https://google.com"}
				]`),
				contentType: "application/json",
				method:      http.MethodPost,
				request:     "/api/shorten/batch",
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.request, bytes.NewReader(tt.args.body))
			request.Header.Set("Content-Type", tt.args.contentType)
			w := httptest.NewRecorder()

			cfg := &config.Config{
				RunAddr:         ":8080",
				BaseAddr:        "http://localhost:8080",
				LogLevel:        "debug",
				FileStoragePath: "/storage.json",
			}
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockURLRepository(ctrl)
			svc := service.NewURLService(repo, cfg)
			h := NewHandler(svc, cfg)

			if tt.name == "Batch Post Success" {
				repo.EXPECT().
					GetOriginalLink(gomock.Any()).
					Return("", repository.ErrRecordNotFound).
					AnyTimes()

				repo.EXPECT().
					SaveBatch(gomock.Any()).
					Return(nil)
			}

			hh := http.HandlerFunc(h.HandleAPIShortenBatch)
			hh.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			if tt.want.code == http.StatusCreated {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}
