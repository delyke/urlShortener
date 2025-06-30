package handler

import "net/http"

type ShortenURLHandler interface {
	HandlePost(w http.ResponseWriter, r *http.Request)
	HandleGet(w http.ResponseWriter, r *http.Request)
}
