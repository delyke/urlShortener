package app

import (
	"bytes"
	"compress/gzip"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompression(t *testing.T) {
	type want struct {
		code               int
		body               string
		isMustBeCompressed bool
	}

	tests := []struct {
		name                      string
		body                      []byte
		contentType               string
		method                    string
		request                   string
		isNeedCompressRequestBody bool
		want                      want
	}{
		{
			name:                      "Send gzip compressed body on POST /api/shorten ContentType text/html",
			body:                      []byte(``),
			contentType:               "text/html",
			method:                    http.MethodPost,
			request:                   "/api/shorten",
			isNeedCompressRequestBody: false,
			want: want{
				code:               http.StatusBadRequest,
				body:               `{"error": "Content-Type must be application/json"}`,
				isMustBeCompressed: true,
			},
		},
		{
			name:                      "Send gzip compressed body on POST /api/shorten ContentType application/json",
			body:                      []byte(``),
			contentType:               "application/json",
			method:                    http.MethodPost,
			request:                   "/api/shorten",
			isNeedCompressRequestBody: false,
			want: want{
				code:               http.StatusBadRequest,
				body:               `{"error": "JSON parse error"}`,
				isMustBeCompressed: true,
			},
		},
		{
			name:                      "Send not gzip answer on POST /api/shorten",
			body:                      []byte(``),
			contentType:               "text/plain",
			method:                    http.MethodPost,
			request:                   "/api/shorten",
			isNeedCompressRequestBody: false,
			want: want{
				code:               http.StatusBadRequest,
				body:               `{"error": "Content-Type must be application/json"}`,
				isMustBeCompressed: false,
			},
		},
		{
			name:                      "Backend can accept gzip compressed body on /api/shorten",
			body:                      []byte(`{"location":"https://www.google.com"}`),
			contentType:               "application/json",
			method:                    http.MethodPost,
			request:                   "/api/shorten",
			isNeedCompressRequestBody: true,
			want: want{
				code:               http.StatusBadRequest,
				body:               `{"error": "URL can't be empty"}`,
				isMustBeCompressed: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request *http.Request

			if tt.isNeedCompressRequestBody {
				var compressedBuf bytes.Buffer
				gzipWriter := gzip.NewWriter(&compressedBuf)
				_, err := gzipWriter.Write(tt.body)
				require.NoError(t, err)
				require.NoError(t, gzipWriter.Close())
				request = httptest.NewRequest(tt.method, tt.request, &compressedBuf)
				request.Header.Set("Content-Encoding", "gzip")
			} else {
				request = httptest.NewRequest(tt.method, tt.request, bytes.NewReader(tt.body))
			}
			request.Header.Set("Accept-Encoding", "gzip")
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			cfg := &config.Config{
				RunAddr:         ":8080",
				BaseAddr:        "http://localhost:8080",
				LogLevel:        "debug",
				FileStoragePath: "storage.json",
			}
			repo, err := repository.NewFileRepository(cfg.FileStoragePath)
			if err != nil {
				t.Error("Failed to initialize repo: ", err)
				return
			}
			svc := service.NewURLService(repo)
			h := handler.NewHandler(svc, cfg)
			l, err := logger.Initialize(cfg.LogLevel)
			if err != nil {
				t.Errorf("Failed to initialize logger: %v", err)
				return
			}
			r := NewRouter(h, l)
			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			require.Equal(t, tt.want.code, result.StatusCode)
			var b []byte
			if tt.want.isMustBeCompressed {
				require.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
				zr, err := gzip.NewReader(result.Body)
				require.NoError(t, err)
				b, err = io.ReadAll(zr)
				require.NoError(t, err)
				require.JSONEq(t, tt.want.body, string(b))
			} else {
				b, err = io.ReadAll(result.Body)
				require.NoError(t, err)
			}
			require.JSONEq(t, tt.want.body, string(b))
		})
	}
}
