package metricserver

import (
	"errors"
	"net/http"
	"strings"

	"github.com/AndreyVLZ/metrics/internal/storage"
)

type metricServer struct {
	store storage.Storage
	mux   *http.ServeMux
}

func New(store storage.Storage) *metricServer {
	srv := &metricServer{
		store: store,
		mux:   http.NewServeMux(),
	}

	srv.mux.HandleFunc("/update/", srv.updateHandler)

	return srv
}

func (s *metricServer) Start() error {
	return http.ListenAndServe("localhost:8080", s.mux)
}

func (s *metricServer) updateHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Only Post", http.StatusBadRequest)
		return
	}

	var err error

	arrPath, err := parseURLPath(r.URL.Path)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	err = s.store.Set(arrPath[0], arrPath[1], arrPath[2])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}

func parseURLPath(path string) ([]string, error) {
	arrPath := strings.Split(path[1:], "/")

	if len(arrPath) != 4 {
		return nil, errors.New("no corect url path")
	}

	return arrPath[1:4], nil
}
