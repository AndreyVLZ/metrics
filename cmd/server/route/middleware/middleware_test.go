package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethod(t *testing.T) {
	testCase := []struct {
		name     string
		m1       string
		m2       string
		wantCode int
	}{
		{
			name:     "positive #1",
			m1:       http.MethodGet,
			m2:       http.MethodGet,
			wantCode: http.StatusOK,
		},
		{
			name:     "positive #2",
			m1:       http.MethodPost,
			m2:       http.MethodPost,
			wantCode: http.StatusOK,
		},
		{
			name:     "negative #1",
			m1:       http.MethodGet,
			m2:       http.MethodPost,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "negative #2",
			m1:       http.MethodPost,
			m2:       http.MethodGet,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			nextHandler := func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}

			req := httptest.NewRequest(test.m1, "/update/", nil)
			rec := httptest.NewRecorder()

			methodHandler := Method(test.m2, nextHandler)
			methodHandler.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, test.wantCode)
		})
	}
}
