package handler

import (
	"context"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

type fakeSrv struct {
	err        error
	mJSON      model.MetricJSON
	arrMetJSON []model.MetricJSON
}

func (fsrv fakeSrv) Update(_ context.Context, met model.MetricJSON) (model.MetricJSON, error) {
	if fsrv.err != nil {
		return model.MetricJSON{}, fsrv.err
	}

	return met, nil
}

func (fsrv fakeSrv) Get(_ context.Context, info model.Info) (model.MetricJSON, error) {
	if fsrv.err != nil {
		return model.MetricJSON{}, fsrv.err
	}

	return fsrv.mJSON, nil
}

func (fsrv fakeSrv) List(_ context.Context) ([]model.MetricJSON, error) {
	if fsrv.err != nil {
		return nil, fsrv.err
	}

	return fsrv.arrMetJSON, nil
}

func (fsrv fakeSrv) AddBatch(_ context.Context, arr []model.MetricJSON) error {
	if fsrv.err != nil {
		return fsrv.err
	}

	return nil
}

func (fsrv fakeSrv) Ping() error {
	if fsrv.err != nil {
		return fsrv.err
	}

	return nil
}

func TestPostJSONUpdateHandle(t *testing.T) {
	type testCase struct {
		name   string
		body   io.Reader
		status int
		header string
		srv    srvUpdater
	}

	tc := []testCase{
		{
			name: "ok count",
			body: strings.NewReader(
				`{"id":"a123","type":"counter","delta":100}`,
			),
			status: http.StatusOK,
			header: ApplicationJSONConst,
			srv:    fakeSrv{},
		},
		{
			name: "ok gauge",
			body: strings.NewReader(
				`{"id":"a123","type":"gauge","value":10}`,
			),
			status: http.StatusOK,
			header: ApplicationJSONConst,
			srv:    fakeSrv{},
		},
		{
			name: "no valid body",
			body: strings.NewReader(
				`{{{`,
			),
			status: http.StatusBadRequest,
			header: ApplicationJSONConst,
			srv:    fakeSrv{},
		},
		{
			name: "srv err",
			body: strings.NewReader(
				`{"id":"a123","type":"counter","delta":100}`,
			),
			status: http.StatusBadRequest,
			header: ApplicationJSONConst,
			srv:    fakeSrv{err: errors.New("srv error")},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			log := slog.Default()
			srv := test.srv

			req := httptest.NewRequest(
				http.MethodPost, "/update/",
				test.body,
			).WithContext(ctx)

			h := PostJSONUpdateHandle(srv, log)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t, test.status, res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return
			}
			// проверяем content-type
			ct := res.Header.Get("Content-Type")
			assert.Equal(t, test.header, ct)
		})
	}
}

func TestPostUpdateHandle(t *testing.T) {
	type testCase struct {
		name    string
		fnParse func(req *http.Request) model.MetricStr
		status  int
		header  string
		srv     srvUpdater
	}

	tc := []testCase{
		{
			name: "ok counter",
			fnParse: func(req *http.Request) model.MetricStr {
				return model.MetricStr{
					InfoStr: model.InfoStr{
						Name:  "Count-1",
						MType: "counter",
					},
					Val: "100",
				}
			},
			status: http.StatusOK,
			header: ApplicationJSONConst,
			srv:    fakeSrv{},
		},

		{
			name: "ok gauge",
			fnParse: func(req *http.Request) model.MetricStr {
				return model.MetricStr{
					InfoStr: model.InfoStr{
						Name:  "Gauge-1",
						MType: "gauge",
					},
					Val: "100.001",
				}
			},
			status: http.StatusOK,
			header: ApplicationJSONConst,
			srv:    fakeSrv{},
		},

		{
			name: "err not valide url",
			fnParse: func(req *http.Request) model.MetricStr {
				return model.MetricStr{
					InfoStr: model.InfoStr{
						Name:  "Gauge-1",
						MType: "gauge",
					},
					Val: "",
				}
			},
			status: http.StatusBadRequest,
			srv:    fakeSrv{},
		},

		{
			name: "err srv",
			fnParse: func(req *http.Request) model.MetricStr {
				return model.MetricStr{
					InfoStr: model.InfoStr{
						Name:  "Count-1",
						MType: "counter",
					},
					Val: "100",
				}
			},
			status: http.StatusBadRequest,
			srv:    fakeSrv{err: errors.New("service error")},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			log := slog.Default()
			srv := test.srv

			req := httptest.NewRequest(
				http.MethodPost, "/update/",
				http.NoBody,
			).WithContext(ctx)

			h := PostUpdateHandle(srv, log, test.fnParse)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t, test.status, res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return
			}
			// проверяем content-type
			ct := res.Header.Get("Content-Type")
			assert.Equal(t, test.header, ct)
		})
	}
}

func TestGetValueHandle(t *testing.T) {
	type testCase struct {
		name    string
		url     string
		fnParse func(req *http.Request) model.InfoStr
		status  int
		header  string
		fnSrv   func() srvGetter
	}

	tc := []testCase{
		{
			name: "ok counter",
			url:  "/value/counter/",
			fnParse: func(req *http.Request) model.InfoStr {
				return model.InfoStr{
					Name:  "PollCount",
					MType: "counter",
				}
			},
			status: http.StatusOK,
			header: "text/plain; charset=utf-8",
			fnSrv: func() srvGetter {
				val := int64(100)
				return fakeSrv{
					mJSON: model.MetricJSON{
						ID:    "PollCount",
						MType: "counter",
						Delta: &val,
					},
				}
			},
		},

		{
			name: "ok gauge",
			url:  "/value/gauge/",
			fnParse: func(req *http.Request) model.InfoStr {
				return model.InfoStr{
					Name:  "Alloc",
					MType: "gauge",
				}
			},
			status: http.StatusOK,
			header: "text/plain; charset=utf-8",
			fnSrv: func() srvGetter {
				val := float64(100.001)
				return fakeSrv{
					mJSON: model.MetricJSON{
						ID:    "Alloc",
						MType: "gauge",
						Value: &val,
					},
				}
			},
		},

		{
			name: "err parse url",
			url:  "/value/counter/",
			fnParse: func(req *http.Request) model.InfoStr {
				return model.InfoStr{
					Name:  "",
					MType: "",
				}
			},
			status: http.StatusBadRequest,
			header: "text/plain; charset=utf-8",
			fnSrv: func() srvGetter {
				val := int64(100)
				return fakeSrv{
					mJSON: model.MetricJSON{
						ID:    "Counter-1",
						MType: "counter",
						Delta: &val,
					},
				}
			},
		},
		{
			name: "err srv",
			url:  "/value/counter/",
			fnParse: func(req *http.Request) model.InfoStr {
				return model.InfoStr{
					Name:  "PollCount",
					MType: "counter",
				}
			},
			status: http.StatusNotFound,
			header: "text/plain; charset=utf-8",
			fnSrv: func() srvGetter {
				return fakeSrv{
					err: errors.New("err srv"),
				}
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			log := slog.Default()
			srv := test.fnSrv()

			req := httptest.NewRequest(
				http.MethodGet, test.url,
				http.NoBody,
			).WithContext(ctx)

			h := GetValueHandle(srv, log, test.fnParse)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t, test.status, res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return
			}
			// проверяем content-type
			ct := res.Header.Get("Content-Type")
			assert.Equal(t, test.header, ct)
		})
	}
}

func TestPostValueHandle(t *testing.T) {
	type testCase struct {
		name   string
		body   io.Reader
		status int
		header string
		fnSrv  func() srvGetter
	}

	tc := []testCase{
		{
			name: "counter ok",
			body: strings.NewReader(
				`{"id":"PollCount","type":"counter","delta":100}`,
			),
			status: http.StatusOK,
			header: ApplicationJSONConst,
			fnSrv: func() srvGetter {
				val := int64(100)
				return fakeSrv{
					mJSON: model.MetricJSON{
						ID:    "PollCount",
						MType: "counter",
						Delta: &val,
					},
				}
			},
		},

		{
			name: "gauge ok",
			body: strings.NewReader(
				`{"id":"Alloc","type":"gauge","value":10.01}`,
			),
			status: http.StatusOK,
			header: ApplicationJSONConst,
			fnSrv: func() srvGetter {
				val := float64(100.001)
				return fakeSrv{
					mJSON: model.MetricJSON{
						ID:    "Alloc",
						MType: "gauge",
						Value: &val,
					},
				}
			},
		},

		{
			name:   "err body not valid",
			body:   strings.NewReader("}}}"),
			status: http.StatusBadRequest,
			header: ApplicationJSONConst,
			fnSrv: func() srvGetter {
				return fakeSrv{}
			},
		},

		{
			name: "err name empty",
			body: strings.NewReader(
				`{"id":"","type":"counter","delta":100}`,
			),
			status: http.StatusBadRequest,
			fnSrv: func() srvGetter {
				return fakeSrv{}
			},
		},

		{
			name: "err mType empty",
			body: strings.NewReader(
				`{"id":"Alloc","type":"","delta":100}`,
			),
			status: http.StatusBadRequest,
			fnSrv: func() srvGetter {
				return fakeSrv{}
			},
		},

		{
			name: "srv custom err",
			body: strings.NewReader(
				`{"id":"PollCount","type":"counter","delta":100}`,
			),
			status: http.StatusNotFound,
			fnSrv: func() srvGetter {
				return fakeSrv{
					err: errors.New("err custom srv"),
				}
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			log := slog.Default()
			srv := test.fnSrv()

			req := httptest.NewRequest(
				http.MethodGet, "/value/",
				test.body,
			).WithContext(ctx)

			h := PostValueHandle(srv, log)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t, test.status, res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return
			}
			// проверяем content-type
			ct := res.Header.Get("Content-Type")
			assert.Equal(t, test.header, ct)
		})
	}
}

func TestPostUpdatesHandler(t *testing.T) {
	type testCase struct {
		name   string
		body   io.Reader
		status int
		srv    srvBatch
	}

	tc := []testCase{
		{
			name: "ok",
			body: strings.NewReader(
				`[{"id":"PollCount","type":"counter","delta":100},{"id":"Alloc","type":"gauge","value":10.01}]`,
			),
			status: http.StatusOK,
			srv:    fakeSrv{},
		},

		{
			name:   "err not valid body",
			body:   strings.NewReader("{{{"),
			status: http.StatusBadRequest,
			srv:    fakeSrv{},
		},

		{
			name: "err srv",
			body: strings.NewReader(
				`[{"id":"PollCount","type":"counter","delta":100},{"id":"Alloc","type":"gauge","value":10.01}]`,
			),
			status: http.StatusBadRequest,
			srv:    fakeSrv{err: errors.New("srv custom err")},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			log := slog.Default()

			req := httptest.NewRequest(
				http.MethodGet, "/value/",
				test.body,
			).WithContext(ctx)

			h := PostUpdatesHandler(test.srv, log)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t, test.status, res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return
			}
		})
	}
}

func TestListHandle(t *testing.T) {
	tpls := `{{define "List"}}{{end}}`
	tplsErr := `{{define "NameTmplErr"}}{{end}}`

	type testCase struct {
		name   string
		body   io.Reader
		status int
		header string
		tmpl   *template.Template
		fnSrv  func() srvBatch
	}

	tc := []testCase{
		{
			name:   "ok",
			status: http.StatusOK,
			header: TextHTMLConst,
			tmpl:   template.Must(template.New("metrics").Parse(tpls)),
			fnSrv: func() srvBatch {
				val := int64(100)
				val2 := float64(10.01)
				return fakeSrv{
					arrMetJSON: []model.MetricJSON{
						{
							ID:    "PollCount",
							MType: "counter",
							Delta: &val,
						},
						{
							ID:    "Alloc",
							MType: "gauge",
							Value: &val2,
						},
					},
				}
			},
		},

		{
			name:   "err srv",
			status: http.StatusBadRequest,
			tmpl:   template.Must(template.New("metrics").Parse(tpls)),
			fnSrv: func() srvBatch {
				return fakeSrv{err: errors.New("err srv.List")}
			},
		},

		{
			name:   "err exec template",
			status: http.StatusInternalServerError,
			tmpl:   template.Must(template.New("metrics").Parse(tplsErr)),
			fnSrv: func() srvBatch {
				return fakeSrv{}
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			log := slog.Default()
			srv := test.fnSrv()

			req := httptest.NewRequest(
				http.MethodGet, "/",
				test.body,
			).WithContext(ctx)

			h := ListHandle(srv, test.tmpl, log)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t, test.status, res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return
			}

			// проверяем content-type
			ct := res.Header.Get("Content-Type")
			assert.Equal(t, test.header, ct)
		})
	}
}

func TestPingHandler(t *testing.T) {
	type testCase struct {
		name   string
		status int
		fnSrv  func() srvPing
	}

	tc := []testCase{
		{
			name:   "ok",
			status: http.StatusOK,
			fnSrv: func() srvPing {
				return fakeSrv{}
			},
		},
		{
			name:   "err",
			status: http.StatusInternalServerError,
			fnSrv: func() srvPing {
				return fakeSrv{err: errors.New("err srv.Ping")}
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			log := slog.Default()
			srv := test.fnSrv()

			req := httptest.NewRequest(
				http.MethodGet, "/ping",
				http.NoBody,
			).WithContext(ctx)

			h := PingHandler(srv, log)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t, test.status, res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return
			}
		})
	}
}
