package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzip(t *testing.T) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("TEST SFS"))
		/*
			valH := rw.Header().Get("Content-Encoding")
			if valH != "gzip" {
				t.Errorf("no header %v\n", valH)
			}
		*/
	})

	t.Run("gzip", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/test",
			strings.NewReader("bodagjjdjfjnjn"),
		)

		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Encoding", "gzip")
		handTest := Gzip(nextHandler)

		ht := httptest.NewRecorder()
		handTest.ServeHTTP(ht, req)

		res := ht.Result()

		fmt.Printf("res %v\n", res.StatusCode)
	})
}
