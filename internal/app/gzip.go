package app

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
)

type compressWriter struct {
	http.ResponseWriter
	buf     bytes.Buffer
	gzw     *gzip.Writer
	written bool
	status  int
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *compressWriter) Write(data []byte) (int, error) {
	if !c.written {
		c.written = true

		if c.status == http.StatusTemporaryRedirect || c.status == http.StatusMovedPermanently {
			c.ResponseWriter.WriteHeader(c.status)
			return c.ResponseWriter.Write(data)
		}

		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		c.ResponseWriter.WriteHeader(c.status)
		c.gzw = gzip.NewWriter(c.ResponseWriter)
	}
	return c.gzw.Write(data)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.status = statusCode
}

func (c *compressWriter) Close() error {
	if c.gzw != nil {
		return c.gzw.Close()
	}
	return nil
}

type compressReader struct {
	r  io.Reader
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c compressReader) Close() error {
	if err := c.zr.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
