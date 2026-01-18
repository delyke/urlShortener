package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/service"
)

func signTestToken(secret string, uid int64) (string, error) {
	c := Claims{
		UserID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return tk.SignedString([]byte(secret))
}

func TestAuthMiddleware_Returns401_WhenTokenHasNoUserID(t *testing.T) {
	cfg := &config.Config{
		HMACSecret:     "test-secret",
		AuthCookieName: "Authorization",
	}

	log, err := logger.Initialize("debug")
	if err != nil {
		t.Fatalf("logger init error: %v", err)
	}
	defer log.Sync()

	var svc *service.URLService = nil

	mw := authCookieMiddleware(cfg, log, svc)

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("final handler must not be called on 401 path")
	})

	token, err := signTestToken(cfg.HMACSecret, 0)
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/any", nil)
	req.AddCookie(&http.Cookie{
		Name:  cfg.AuthCookieName,
		Value: token,
		Path:  "/",
	})
	rr := httptest.NewRecorder()

	mw(final).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
