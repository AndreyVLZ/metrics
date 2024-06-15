package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeStore struct {
	storage
}

func TestPing(t *testing.T) {
	storage := &fakeStore{}

	adapter := Ping(storage)

	err := adapter.Ping()
	assert.NoError(t, err)
}
