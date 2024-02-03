package mainhandler

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AndreyVLZ/metrics/cmd/server/urlpath"
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

type want struct {
	contentType string
	statusCode  int
}

type fakeStore struct {
	storage.Storage
	err error
	m   metric.MetricDB
}

func (fs fakeStore) Get(ctx context.Context, m metric.MetricDB) (metric.MetricDB, error) {
	if fs.err != nil {
		return metric.MetricDB{}, fs.err
	}
	return fs.m, nil
}

func (fs fakeStore) Set(ctx context.Context, m metric.MetricDB) (metric.MetricDB, error) {
	if fs.err != nil {
		return metric.MetricDB{}, fs.err
	}
	return fs.m, nil
}

func (fs fakeStore) List(ctx context.Context) []metric.MetricDB {
	if fs.err != nil {
		return []metric.MetricDB{}
	}
	return []metric.MetricDB{fs.m}
}

type fakeEmbed struct {
	err    error
	metric metric.MetricDB
}

func (e fakeEmbed) GetMetricDBFromRequest(*http.Request) (metric.MetricDB, error) {
	if e.err != nil {
		return metric.MetricDB{}, e.err
	}
	return e.metric, nil
}

func (e fakeEmbed) GetUpdateMetricDBFromRequest(*http.Request) (metric.MetricDB, error) {
	if e.err != nil {
		return metric.MetricDB{}, e.err
	}
	return e.metric, nil
}

func TestListHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	tc := map[string]struct {
		tName   string
		tmpls   *template.Template
		store   storage.Storage
		request string
		want    want
	}{
		"positive": {
			tName: "#1",
			tmpls: template.Must(template.New("metrics").Parse(tpls)),
			store: fakeStore{
				Storage: memstorage.New(),
				m: metric.NewMetricDB(
					"MyCounter",
					metric.Counter(1),
				),
			},
			request: "/",
			want: want{
				contentType: TextHTMLConst,
				statusCode:  http.StatusOK,
			},
		},

		"negative": {
			tName: "err execute html temlate",
			tmpls: template.Must(template.New("metrics").Parse(tpls)),
			store: fakeStore{
				Storage: memstorage.New(),
			},
			request: "/",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
			},
		},
	}

	for nTest, test := range tc {
		t.Run(nTest+test.tName, func(t *testing.T) {
			mh := mainHandlers{
				tmpls:            test.tmpls,
				store:            test.store,
				EmbedingHandlers: nil,
			}
			request := httptest.NewRequest(http.MethodGet, test.request, http.NoBody)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(mh.ListHandler().ServeHTTP)
			h(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}

func TestGetValueHandler(t *testing.T) {
	tc := map[string][]struct {
		tName   string
		store   storage.Storage
		embed   EmbedingHandlers
		want    want
		expBody string
	}{
		"positive": {
			{
				tName: "#1",
				store: fakeStore{
					m: metric.NewMetricDB(
						"myCounter",
						metric.Counter(1),
					),
				},
				embed:   fakeEmbed{},
				expBody: "1",
				want: want{
					statusCode:  http.StatusOK,
					contentType: "text/plain; charset=utf-8",
				},
			},
		},

		"negative": {
			{
				tName:   "#1 [store error]",
				store:   fakeStore{err: memstorage.ErrNotSupportedType},
				embed:   fakeEmbed{},
				expBody: "",
				want: want{
					statusCode:  http.StatusNotFound,
					contentType: "text/plain; charset=utf-8",
				},
			},

			{
				tName:   "#2 [error empty name field]",
				store:   fakeStore{err: memstorage.ErrNotSupportedType},
				embed:   fakeEmbed{err: urlpath.ErrEmptyNameField},
				expBody: "",
				want: want{
					statusCode:  http.StatusNotFound,
					contentType: "text/plain; charset=utf-8",
				},
			},

			{
				tName:   "#3 [error no correct url]",
				store:   fakeStore{err: memstorage.ErrNotSupportedType},
				embed:   fakeEmbed{err: urlpath.ErrNoCorrectURLPath},
				expBody: "",
				want: want{
					statusCode:  http.StatusBadRequest,
					contentType: "text/plain; charset=utf-8",
				},
			},
		},
	}

	for nTest, tests := range tc {
		for _, test := range tests {
			mh := New(test.store)
			mh.EmbedingHandlers = test.embed
			t.Run(nTest+" "+test.tName, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				w := httptest.NewRecorder()
				h := http.HandlerFunc(mh.GetValueHandler().ServeHTTP)
				h(w, request)
				result := w.Result()

				assert.Equal(t, test.want.statusCode, result.StatusCode)
				assert.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))
				if test.expBody != "" {
					body, _ := io.ReadAll(result.Body)
					defer result.Body.Close()
					assert.Equal(t, test.expBody, string(body))
				}
			})
		}
	}
}

func TestPostValueHandler(t *testing.T) {
	tc := map[string][]struct {
		tName   string
		store   storage.Storage
		want    want
		inBody  string
		expBody string
		err     error
	}{

		"positive": {
			{
				tName:  "#1 [counter]",
				inBody: `{"id":"myCounter","type":"counter"}`,
				store: fakeStore{
					m: metric.NewMetricDB(
						"myCounter",
						metric.Counter(1),
					),
				},
				expBody: `{"id":"myCounter","type":"counter","delta":1}
`,
				want: want{
					statusCode:  http.StatusOK,
					contentType: ApplicationJSONConst,
				},
			},

			{
				tName:  "#2 [gauge]",
				inBody: `{"id":"myGauge","type":"gauge"}`,
				store: fakeStore{
					m: metric.NewMetricDB(
						"myGauge",
						metric.Gauge(1.1),
					),
				},
				expBody: `{"id":"myGauge","type":"gauge","value":1.1}
`,
				want: want{
					statusCode:  http.StatusOK,
					contentType: ApplicationJSONConst,
				},
			},
		},

		"negative": {
			{
				tName:   "#1 [in body nil]",
				inBody:  ``,
				store:   nil,
				expBody: ErrJSONSyntax.Error() + "\n",
				want: want{
					statusCode:  http.StatusNotFound,
					contentType: "text/plain; charset=utf-8",
				},
			},
			{
				tName:   "#2 [err read from body]",
				inBody:  `{"":}`,
				store:   nil,
				expBody: ErrJSONSyntax.Error() + "\n",
				want: want{
					statusCode:  http.StatusNotFound,
					contentType: "text/plain; charset=utf-8",
				},
			},
			{
				tName:   "#2 [err not type support]",
				inBody:  `{"id":"myCounter","type":"errType"}`,
				store:   nil,
				expBody: ErrTypeNotSupport.Error() + "\n",
				want: want{
					statusCode:  http.StatusNotFound,
					contentType: "text/plain; charset=utf-8",
				},
			},
		},
	}

	for nTest, tests := range tc {
		for _, test := range tests {
			mh := New(test.store)
			t.Run(nTest+" "+test.tName, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.inBody))
				w := httptest.NewRecorder()
				h := http.HandlerFunc(mh.PostValueHandler().ServeHTTP)
				h(w, request)
				result := w.Result()

				assert.Equal(t, test.want.statusCode, result.StatusCode)
				assert.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))
				if test.expBody != "" {
					actualBody, _ := io.ReadAll(result.Body)
					defer result.Body.Close()
					assert.Equal(t, test.expBody, string(actualBody))
				}
			})
		}
	}
}

/*
func TestHandlersList(t *testing.T) {
	tc := map[string][]struct {
		tName   string
		url     string
		method  string
		store   storage.Storage
		want    want
		inBody  string
		expBody string
		err     error
	}{
		"positive": {
			{
				tName:  "#1",
				url:    "/",
				method: http.MethodPost,
				store: fakeStore{
					m: metric.NewMetricDB(
						"myCounter",
						metric.Counter(1),
					),
				},
			},
			{
				tName:  "#2",
				url:    "/value/",
				method: http.MethodGet,
			},
		},
	}

	for nTest, tests := range tc {
		for _, test := range tests {
			t.Run(nTest+" "+test.tName, func(t *testing.T) {
				mh := New(test.store)
				mh.EmbedingHandlers = servemux.NewServeMuxHandle()
				h := http.HandlerFunc(servemux.New().SetHandlers(mh).ServeHTTP)

				request := httptest.NewRequest(
					test.method,
					test.url,
					strings.NewReader(test.inBody),
				)

				w := httptest.NewRecorder()
				h(w, request)
				result := w.Result()

				assert.Equal(t, test.want.statusCode, result.StatusCode)
				//assert.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))
				if test.expBody != "" {
					actualBody, _ := io.ReadAll(result.Body)
					defer result.Body.Close()
					assert.Equal(t, test.expBody, string(actualBody))
				}
			})
		}
	}
}
*/
