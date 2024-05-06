package middleware

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"
)

type iLogger interface {
	Info(msg string, args ...any)
}

type responseData struct {
	status int
	size   int
}

func newResponceData() *responseData {
	return &responseData{
		status: http.StatusOK,
	}
}

func (rd *responseData) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("status", rd.status),
		slog.Int("size", rd.size),
	)
}

type requestData struct {
	uri      string
	method   string
	duration time.Duration
}

func (rd *requestData) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("uri", rd.uri),
		slog.String("method", rd.method),
		slog.Duration("duration", rd.duration),
	)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func newLoggingResponseWriter(rw http.ResponseWriter, resData *responseData) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: rw,
		responseData:   resData,
	}
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.responseData.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func Logging(log iLogger, next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()
		reqData := &requestData{
			uri:    req.URL.String(),
			method: req.Method,
		}

		resData := newResponceData()

		defer func() {
			reqData.duration = time.Since(start)
			log.Info("INFO", "response", resData, "request", reqData)
		}()

		lmw := newLoggingResponseWriter(rw, resData)

		next.ServeHTTP(lmw, req)
	}
}
