package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	var storage storage

	adapter := Ping(storage)

	err := adapter.Ping()
	assert.NoError(t, err)
}
