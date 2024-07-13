package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	fnSetHeader func(req *http.Request)
	name        string
	statusCode  int
}

func fnCheck(t *testing.T, tName string, val string, fn func(http.Handler) http.Handler) {
	tc := []testCase{
		{
			name:       "ok",
			statusCode: http.StatusOK,
			fnSetHeader: func(req *http.Request) {
				req.Header.Add("Content-Type", val)
			},
		},

		{
			name:        "not header",
			statusCode:  http.StatusUnsupportedMediaType,
			fnSetHeader: func(req *http.Request) {},
		},
	}

	for _, test := range tc {
		t.Run(tName+" "+test.name, func(t *testing.T) {
			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(
				http.MethodPost,
				"/test",
				http.NoBody,
			)

			test.fnSetHeader(req)

			ht := httptest.NewRecorder()
			handler := fn(next)

			handler.ServeHTTP(ht, req)

			res := ht.Result()

			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, test.statusCode, res.StatusCode)
		})
	}
}

func TestAppJson(t *testing.T) {
	fnCheck(t, "appJson", appJSON, AppJSON())
}

func TestTextPlain(t *testing.T) {
	fnCheck(t, "textPlain", textPlain, TextPlain())
}
