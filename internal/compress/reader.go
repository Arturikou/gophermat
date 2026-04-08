package compress

import (
	"compress/gzip"
	"io"
)

type Reader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewReader(r io.ReadCloser) (*Reader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &Reader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *Reader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *Reader) Close() error {
	if err := c.zr.Close(); err != nil {
		return err
	}
	return c.r.Close()
}
