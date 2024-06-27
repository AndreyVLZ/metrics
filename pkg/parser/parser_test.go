package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	acFile := "123"
	exFile := ""
	acVal := "acVal"
	exVal := "exVal"

	File(&acFile,
		func(strPtr *string) error {
			*strPtr = exFile

			return nil
		},
	)

	Value(&acVal,
		func(strPtr *string) error {
			*strPtr = exVal

			return nil
		},
	)

	if err := Parse([]string{}); err != nil {
		t.Errorf("parse: %v\n", err)
	}

	assert.Equal(t, exFile, acFile)
	assert.Equal(t, exVal, acVal)
}
