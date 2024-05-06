package service

import (
	"errors"
	"testing"
)

func (fs *fakeStore) Ping() error {
	if fs.err != nil {
		return fs.err
	}

	return nil
}

func TestPing(t *testing.T) {
	type testCase struct {
		name      string
		fakeStore *fakeStore
		isErr     bool
	}

	tc := []testCase{
		{
			name:      "ok",
			fakeStore: &fakeStore{},
			isErr:     false,
		},

		{
			name:      "ping err",
			fakeStore: &fakeStore{err: errors.New("ping err")},
			isErr:     false,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			srv := New(test.fakeStore)
			err := srv.Ping()

			if test.isErr && err != nil {
				t.Error("want err")
			}
		})
	}
}
