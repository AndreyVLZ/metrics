package middleware

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/hash"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	type testCase struct {
		name       string
		key        string
		hKey       string
		statusCode int
	}

	var secret = "SECRET-KEY"

	tc := []testCase{
		{
			name:       "valid key",
			key:        secret,
			hKey:       secret,
			statusCode: http.StatusOK,
		},

		{
			name:       "not valid key",
			key:        "S",
			hKey:       secret,
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "key empty",
			key:        secret,
			hKey:       "",
			statusCode: http.StatusOK,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			testBody := "TEST STRING"

			req := httptest.NewRequest(
				http.MethodPost,
				"/test",
				strings.NewReader(testBody),
			)

			// Вычисляет хеш и передавает его в HTTP-заголовке
			sum, err := hash.SHA256([]byte(testBody), []byte(test.key))
			if err == nil {
				req.Header.Set("HashSHA256", hex.EncodeToString(sum))
			}

			nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				if _, err := rw.Write([]byte("write data")); err != nil {
					t.Fatalf("readBody:%v\n", err)
				}
			})

			ht := httptest.NewRecorder()
			handler := Hash(test.hKey, nextHandler)

			handler.ServeHTTP(ht, req)

			res := ht.Result()

			assert.Equal(t, test.statusCode, res.StatusCode)
		})
	}
}
