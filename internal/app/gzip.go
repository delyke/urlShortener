package app

import (
	"io"
	"net/http"
	"strings"
)

// gzipWriter wraps the response writer to selectively gzip responses.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write compresses supported content types and forwards others untouched.
func (g *gzipWriter) Write(data []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
		return g.Writer.Write(data)
	}
	return g.ResponseWriter.Write(data)
}
