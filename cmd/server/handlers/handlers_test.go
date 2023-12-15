package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/AndreyVLZ/metrics/cmd/server/middleware"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	type want struct {
		code int
		//response    string
		contentType string
	}

	tests := []struct {
		name   string
		path   string
		method string
		want   want
	}{
		{
			name:   "positive test #1",
			method: http.MethodPost,
			path:   "/update/counter/myCounter/10",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name:   "negative #1 method GET",
			method: http.MethodGet,
			path:   "/update/counter/myCounter/10",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "negative #2 bad url",
			method: http.MethodPost,
			path:   "/update/counter/myCounter",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "negative #3 no support type",
			method: http.MethodPost,
			path:   "/update/co/myCounter/10",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			mh := NewMetricHandler(memstorage.New(
				memstorage.NewGaugeRepo(),
				memstorage.NewCounterRepo(),
			))

			handler := http.HandlerFunc(
				middleware.ContentType("text/plain",
					middleware.Method(http.MethodPost, mh.UpdateHandler),
				),
			)
			handler.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			//resBody, err := io.ReadAll(res.Body)

			//require.NoError(t, err)
			//assert.JSONEq(t, test.want.response, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestParseURLPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		res  []string
		err  error
	}{
		{
			name: "ok",
			path: "update/counter/MyCounter/10",
			res:  []string{"counter", "MyCounter", "10"},
			err:  nil,
		},
		{
			name: "false",
			path: "update/counter/MyCounter",
			res:  nil,
			err:  ErrNoCorrectURLPath,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			got, exErr := parseURLPath(test.path, 4)
			if !errors.Is(test.err, exErr) {
				t.Errorf("ERR parseURLPath() = %v, want %v", exErr, test.err)
			}
			if !reflect.DeepEqual(got, test.res) {
				t.Errorf("parseURLPath() = %v, want %v", got, test.res)
			}
		})
	}
}
