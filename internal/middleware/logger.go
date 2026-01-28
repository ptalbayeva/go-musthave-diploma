package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log = zap.NewNop()

type (
	responseData struct {
		status int
		size   int
	}

	loggerResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl

	return nil
}

func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			start := time.Now()
			uri := req.RequestURI
			method := req.Method
			data := &responseData{
				size:   0,
				status: 0,
			}

			lw := loggerResponseWriter{
				ResponseWriter: rw,
				responseData:   data,
			}

			next.ServeHTTP(&lw, req)

			duration := time.Since(start)

			Log.Info("New Request:",
				zap.String("uri", uri),
				zap.String("method", method),
				zap.Duration("duration", duration),
				zap.Int("size", lw.responseData.size),
				zap.Int("status", lw.responseData.status),
			)
		})
	}
}

func (r loggerResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	return size, err
}

func (r loggerResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
