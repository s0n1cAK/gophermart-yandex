package server

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := loggerRW{
				ResponseWriter: w,
				responseData: &responseData{
					status: 0,
					size:   0,
				},
			}

			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			logger.Info("",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", lw.responseData.status),
				zap.Duration("duration", duration),
				zap.Int("size", lw.responseData.size),
			)
		}
		return http.HandlerFunc(logFn)
	}
}

func gzipCompession() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		compressFn := func(w http.ResponseWriter, r *http.Request) {

			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				w.Header().Set("Content-Encoding", "gzip")
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			h.ServeHTTP(ow, r)

		}
		return http.HandlerFunc(compressFn)
	}
}
