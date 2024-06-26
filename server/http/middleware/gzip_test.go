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

		if err := res.Body.Close(); err != nil {
			t.Error(err)
		}

		fmt.Printf("res %v\n", res.StatusCode)
	})
}

func BenchmarkGzip(b *testing.B) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("TEST SFS"))
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		strings.NewReader("bodagjjdjfjnjn"),
	)

	for i := 0; i < b.N; i++ {
		handTest := Gzip(nextHandler)
		ht := httptest.NewRecorder()
		handTest.ServeHTTP(ht, req)

		res := ht.Result()

		if err := res.Body.Close(); err != nil {
			b.Errorf("boby close err: %v\n", err)
		}

		if res.StatusCode != http.StatusOK {
			b.Errorf("status: %d\n", res.StatusCode)
		}
	}
}
