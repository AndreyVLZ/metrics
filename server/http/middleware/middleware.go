package middleware

import (
	"net/http"
)

const (
	appJSON   string = "application/json"
	textPlain string = "text/plain"
)

type Middle func(http.Handler) http.Handler

func AppJSON() Middle   { return contentType(appJSON) }
func TextPlain() Middle { return contentType(textPlain) }

func contentType(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Content-Type") != contentType {
				rw.WriteHeader(http.StatusUnsupportedMediaType)

				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}
