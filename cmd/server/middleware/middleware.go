package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
)

var ErrOnlyPostRequest error = errors.New("only Post")

type (
	requestData struct {
		uri      string
		method   string
		duration time.Duration
	}

	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func newResponceData() *responseData {
	return &responseData{
		status: http.StatusOK,
	}
}

func (rd *requestData) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("uri", rd.uri),
		slog.String("method", rd.method),
		slog.Duration("duration", rd.duration),
	)
}

func (rd *responseData) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("status", rd.status),
		slog.Int("size", rd.size),
	)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)

	r.responseData.size += size

	return size, err
}

func newLoggingResponseWriter(rw http.ResponseWriter, resData *responseData) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: rw,
		responseData:   resData,
	}
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.responseData.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func Logging(log *slog.Logger, next http.Handler) http.HandlerFunc {
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

func Method(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, ErrOnlyPostRequest.Error(), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(rw, req)
	}
}

func ContentType(ct string, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", ct)
		next.ServeHTTP(rw, req)
	}
}
