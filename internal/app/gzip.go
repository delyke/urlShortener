package app

import (
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
		return g.Writer.Write(data)
	}
	return g.ResponseWriter.Write(data)
}
