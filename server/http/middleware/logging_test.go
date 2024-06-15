package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeLog struct {
	args []any
}

func (fl *fakeLog) Info(msg string, args ...any) {
	fl.args = args
}

func TestLogging(t *testing.T) {
	testBody := "TEST STRING"

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		strings.NewReader(testBody),
	)
	var wantLenBytes int
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var err error
		rw.WriteHeader(http.StatusOK)
		wantLenBytes, err = rw.Write([]byte("write data"))
		if err != nil {
			t.Fatalf("readBody:%v\n", err)
		}
	})

	flog := &fakeLog{}

	ht := httptest.NewRecorder()

	handler := Logging(flog, nextHandler)

	handler.ServeHTTP(ht, req)

	if len(flog.args) != 4 {
		t.Errorf("err len: want [%d], actual [%d]\n", 4, len(flog.args))
	}

	resDataAny := flog.args[1]
	reqDataAny := flog.args[3]

	resData, ok := resDataAny.(*responseData)
	if !ok {
		t.Fatalf("build resData\n")

	}

	reqData, ok := reqDataAny.(*requestData)
	if !ok {
		t.Fatalf("build reqData\n")
	}

	assert.Equal(t, http.MethodPost, reqData.method)
	assert.Equal(t, "/test", reqData.uri)
	assert.Equal(t, wantLenBytes, resData.size)
	assert.Equal(t, http.StatusOK, resData.status)
}
