package middleware

import (
	"net/http"
	"strings"

	"github.com/Arturikou/internal/compress"
)

func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := compress.NewReader(r.Body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			r.Body = cr
		}

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if !supportsGzip {
			next.ServeHTTP(w, r)
			return
		}
		cw := compress.NewWriter(w)
		defer cw.Close()

		next.ServeHTTP(cw, r)
	})
}
