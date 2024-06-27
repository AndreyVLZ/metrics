package field

import (
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
	"github.com/stretchr/testify/assert"
)

func initData(key string, val interface{}) { myField.data[key] = val }

func clearData() {
	myField.data = make(map[string]interface{}, 0)
}

func TestDuration(t *testing.T) {
	myField = &field{
		data: make(map[string]interface{}, 0),
	}

	fieldName := "fName"

	t.Run("ok", func(t *testing.T) {
		t.Cleanup(func() {
			clearData()
		})

		valStr := "13s"

		initData(fieldName, valStr)

		exVal, err := time.ParseDuration(valStr)
		if err != nil {
			t.Fatalf("parse duration: %v", err)
		}

		var acVal time.Duration
		if err := Duration(fieldName)(&acVal); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, exVal, acVal)
	})

	t.Run("err parse", func(t *testing.T) {
		t.Cleanup(func() {
			clearData()
		})

		valStr := "13"

		initData(fieldName, valStr)

		var acVal time.Duration
		if err := Duration(fieldName)(&acVal); err == nil {
			t.Fatal("want err")
		}
	})

	t.Run("err not set", func(t *testing.T) {
		t.Cleanup(func() {
			clearData()
		})

		var acVal time.Duration
		err := Duration(fieldName)(&acVal)
		assert.Equal(t, perr.ErrNotSet, err)
	})
}

func TestBool(t *testing.T) {
	fieldName := "fName"

	t.Run("ok", func(t *testing.T) {
		t.Cleanup(func() {
			clearData()
		})

		exVal := true

		initData(fieldName, exVal)

		var acVal bool
		if err := Bool(fieldName)(&acVal); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, exVal, acVal)
	})

	t.Run("err not set", func(t *testing.T) {
		t.Cleanup(func() {
			clearData()
		})

		var acVal bool
		err := Bool(fieldName)(&acVal)
		assert.Equal(t, perr.ErrNotSet, err)
	})
}

func TestString(t *testing.T) {
	fieldName := "fName"

	t.Run("ok", func(t *testing.T) {
		t.Cleanup(func() {
			clearData()
		})

		exVal := "value"

		initData(fieldName, exVal)

		var acVal string
		if err := String(fieldName)(&acVal); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, exVal, acVal)
	})

	t.Run("err not set", func(t *testing.T) {
		t.Cleanup(func() {
			clearData()
		})

		var acVal string
		err := String(fieldName)(&acVal)
		assert.Equal(t, perr.ErrNotSet, err)
	})
}
