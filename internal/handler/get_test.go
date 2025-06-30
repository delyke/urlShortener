package handler_test

import (
	"bytes"
	"github.com/delyke/urlShortener/internal/app"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_HandleGet(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name    string
		request string
		method  string
		want    want
	}{
		{
			name:    "Empty string",
			request: `/`,
			method:  http.MethodGet,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name:    "Invalid url",
			request: `/she4894t`,
			method:  http.MethodGet,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name:    "Redirect to shorted Url",
			request: `/`,
			method:  http.MethodGet,
			want: want{
				code:        http.StatusTemporaryRedirect,
				contentType: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := repository.NewLocalRepository()
			svc := service.NewURLService(repo)
			h := handler.NewHandler(svc)
			r := app.NewRouter(h)

			if tt.name == "Redirect to shorted Url" {
				wPost := httptest.NewRecorder()
				postRequest := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("https://yandex.com")))
				r.ServeHTTP(wPost, postRequest)
				postResult := wPost.Result()
				body, _ := io.ReadAll(postResult.Body)
				postResult.Body.Close()
				tt.request = strings.TrimPrefix(string(body), "http://localhost:8080")
			}
			w := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.request, nil)
			r.ServeHTTP(w, request)

			result := w.Result()
			result.Body.Close()

			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
}
