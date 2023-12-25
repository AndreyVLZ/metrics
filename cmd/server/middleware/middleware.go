package middleware

import (
	"errors"
	"net/http"
)

var ErrOnlyPostRequest error = errors.New("only Post")

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
