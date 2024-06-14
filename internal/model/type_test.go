package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportType(t *testing.T) {
	assert.Equal(t, "counter", TypeCountConst.String())
	assert.Equal(t, "gauge", TypeGaugeConst.String())
}

func TestParseType(t *testing.T) {
	t.Run("parse counter ok", func(t *testing.T) {
		mtype, err := ParseType("counter")
		assert.NoError(t, err)
		assert.Equal(t, TypeCountConst, mtype)
	})

	t.Run("parse gauge ok", func(t *testing.T) {
		mtype, err := ParseType("gauge")
		assert.NoError(t, err)
		assert.Equal(t, TypeGaugeConst, mtype)
	})

	t.Run("parse type err", func(t *testing.T) {
		_, err := ParseType("cg")
		errors.Is(err, ErrTypeNotSupport)
		if err == nil {
			t.Error("want err")
		}
	})
}
