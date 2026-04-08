package compress

import (
	"compress/gzip"
	"net/http"
	"strings"
)

var compressibleTypes = []string{
	"application/json",
	"text/html",
}

type Writer struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{
		w: w,
	}
}

func (c *Writer) Header() http.Header {
	return c.w.Header()
}

func (c *Writer) Write(p []byte) (int, error) {
	if c.zw == nil && c.isCompressible() {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length")
		c.zw = gzip.NewWriter(c.w)
	}

	if c.zw != nil {
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

func (c *Writer) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *Writer) Close() error {
	if c.zw != nil {
		return c.zw.Close()
	}
	return nil
}

func (c *Writer) isCompressible() bool {
	contentType := c.w.Header().Get("Content-Type")
	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}
