package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testCase = []struct {
	name   string
	path   string
	method string
	body   string
	code   int
}{
	{
		name:   "positive #1",
		path:   "/update/typeStr/nameVal/val",
		method: http.MethodPost,
		body:   "update",
		code:   http.StatusOK,
	},
	{
		name:   "positive #2",
		path:   "/value/typeStr/nameVal",
		method: http.MethodGet,
		body:   "get value",
		code:   http.StatusOK,
	},
	{
		name:   "positive #3",
		path:   "/",
		method: http.MethodGet,
		body:   "list",
		code:   http.StatusOK,
	},
}

type testHandlers struct{}

func (h *testHandlers) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("update"))
}
func (h *testHandlers) GetValueHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("get value"))
}
func (h *testHandlers) ListHandler(rw http.ResponseWriter, rec *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("list"))
}

func TestSetHandlersMux(t *testing.T) {
	th := &testHandlers{}
	srvMux := NewServeMux()

	srvMux.SetHandlers(th)

	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rec := httptest.NewRecorder()
			srvMux.mux.ServeHTTP(rec, req)

			assert.Equal(t, test.body, rec.Body.String())

			sc := statusCode(t, rec)
			assert.Equal(t, test.code, sc)
		})
	}

}

func TestSetHandlersChi(t *testing.T) {
	th := &testHandlers{}
	chiMux := NewChiMux()

	chiMux.SetHandlers(th)

	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rec := httptest.NewRecorder()
			chiMux.mux.ServeHTTP(rec, req)

			assert.Equal(t, test.body, rec.Body.String())

			sc := statusCode(t, rec)
			assert.Equal(t, test.code, sc)
		})
	}
}

func statusCode(t *testing.T, rec *httptest.ResponseRecorder) int {
	res := rec.Result()
	body := res.Body
	if err := body.Close(); err != nil {
		t.Fatal(err)
	}
	return res.StatusCode
}
