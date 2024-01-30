package mainhandler

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndreyVLZ/metrics/cmd/server/urlpath"
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

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
		rw      http.ResponseWriter
	}{
		"positive": {
			tName: "Get",
			tmpls: template.Must(template.New("metrics").Parse(tpls)),
			store: fakeStore{
				Storage: memstorage.New(),
			},
			rw:      httptest.NewRecorder(),
			request: "/list",
			want: want{
				contentType: TextHTMLConst,
				statusCode:  http.StatusOK,
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
			h := http.HandlerFunc(mh.ListHandler)
			h(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, test.want.statusCode, result.StatusCode)
			assert.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}

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
			mh := NewMainHandlers(
				test.store,
				test.embed,
			)
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
