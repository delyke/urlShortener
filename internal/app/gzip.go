package app

import (
	"compress/gzip"
	"io"
	"net/http"
)

type compressWriter struct {
	w             http.ResponseWriter
	zw            *gzip.Writer
	headerWritten bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(b []byte) (int, error) {
	return c.zw.Write(b)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.headerWritten {
		return
	}
	c.headerWritten = true
	if (statusCode >= 200 && statusCode < 300) || statusCode >= 400 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
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
