package app

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/delyke/urlShortener/internal/app/appctx"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
	"time"
)

func NewRouter(h *handler.Handler, l *logger.Logger, cfg *config.Config, svc *service.URLService) chi.Router {
	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Use(authCookieMiddleware(cfg, l, svc))
	r.Use(l.RequestLogger)
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.HandlePost)
		r.Get("/", h.HandleGet)
		r.Get("/{shortURL}", h.HandleGet)
		r.Post("/api/shorten", h.HandleAPIShorten)
		r.Post("/api/shorten/batch", h.HandleAPIShortenBatch)
		r.Get("/ping", h.HandlePing)
		r.Get("/api/user/urls", h.HandleAPIUserURLs)
	})
	return r
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

func authCookieMiddleware(cfg *config.Config, l *logger.Logger, svc *service.URLService) func(http.Handler) http.Handler {
	cookieName := cfg.AuthCookieName
	if cookieName == "" {
		cookieName = "Authorization"
	}

	secret := cfg.HMACSecret

	sign := func(uid int64) (string, error) {
		claims := Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID: uid,
		}
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return t.SignedString([]byte(secret))
	}

	parse := func(tokenString string) (*Claims, *jwt.Token, error) {
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				l.Warnf("Unexpected signing method: %v", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		return claims, token, err
	}

	setCookieAndAuthHeader := func(w http.ResponseWriter, token string) {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(24 * time.Hour),
		})
		w.Header().Set("Authorization", token)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""
			if c, err := r.Cookie(cookieName); err == nil && c != nil && c.Value != "" {
				tokenString = c.Value
			}
			if tokenString == "" {
				if hdr := r.Header.Get("Authorization"); hdr != "" {
					tokenString = hdr
				}
			}

			if tokenString == "" {
				uid, err := svc.CreateUser()
				if err != nil {
					l.Errorf("Error creating user: %v", err)
					writeJSON(w, http.StatusInternalServerError, &ErrorResponse{Error: "failed to create user"})
					return
				}
				token, err := sign(uid)
				if err != nil {
					l.Errorf("Error signing user: %v", err)
					writeJSON(w, http.StatusInternalServerError, &ErrorResponse{Error: "failed to sign user"})
					return
				}
				setCookieAndAuthHeader(w, token)
				r = r.WithContext(appctx.SetUserID(r.Context(), uid))
				next.ServeHTTP(w, r)
				return
			}

			claims, parsed, err := parse(tokenString)
			if err != nil || !parsed.Valid {
				uid, err := svc.CreateUser()
				if err != nil {
					l.Errorf("Error creating user: %v", err)
					writeJSON(w, http.StatusInternalServerError, &ErrorResponse{Error: "failed to create user"})
					return
				}
				token, err := sign(uid)
				if err != nil {
					l.Errorf("Error signing user: %v", err)
					writeJSON(w, http.StatusInternalServerError, &ErrorResponse{Error: "failed to sign user"})
					return
				}
				setCookieAndAuthHeader(w, token)
				r = r.WithContext(appctx.SetUserID(r.Context(), uid))
				next.ServeHTTP(w, r)
				return
			}

			if claims.UserID == 0 {
				writeJSON(w, http.StatusUnauthorized, &ErrorResponse{Error: "not user id in token"})
				return
			}
			r = r.WithContext(appctx.SetUserID(r.Context(), claims.UserID))
			next.ServeHTTP(w, r)
		})
	}
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Invalid gzip body"})
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		originalWriter := w
		clientAcceptGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if clientAcceptGzip {
			isRelevantContentType := strings.Contains(r.Header.Get("Content-Type"), "application/json") ||
				strings.Contains(r.Header.Get("Content-Type"), "text/html")
			if isRelevantContentType {
				gzWriter := gzip.NewWriter(w)
				defer gzWriter.Close()
				w.Header().Set("Content-Encoding", "gzip")
				originalWriter = &gzipWriter{
					ResponseWriter: w,
					Writer:         gzWriter,
				}
			}
		}
		next.ServeHTTP(originalWriter, r)
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
