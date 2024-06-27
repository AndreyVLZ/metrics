package env

import (
	"os"
	"strconv"
	"testing"

	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	envName := "myENV"

	t.Run("ok", func(t *testing.T) {
		envVal := 777

		t.Cleanup(func() {
			os.Unsetenv(envName)
		})

		if err := os.Setenv(envName, strconv.Itoa(envVal)); err != nil {
			t.Fatalf("set env: %v\n", err)
		}

		var exVal int
		if err := Int(envName)(&exVal); err != nil {
			t.Fatalf("fnInt: %v\n", err)
		}

		assert.Equal(t, exVal, envVal)
	})

	t.Run("err parse", func(t *testing.T) {
		envVal := "x7"

		t.Cleanup(func() {
			os.Unsetenv(envName)
		})

		if err := os.Setenv(envName, envVal); err != nil {
			t.Fatalf("set env: %v\n", err)
		}

		var exVal int
		if err := Int(envName)(&exVal); err == nil {
			t.Fatal("want err")
		}
	})

	t.Run("err not set", func(t *testing.T) {
		var exVal int
		err := Int(envName)(&exVal)
		assert.Equal(t, err, perr.ErrNotSet)
	})
}

func TestBool(t *testing.T) {
	envName := "myENV"

	t.Run("ok", func(t *testing.T) {
		envVal := true

		t.Cleanup(func() {
			os.Unsetenv(envName)
		})

		if err := os.Setenv(envName, strconv.FormatBool(envVal)); err != nil {
			t.Fatalf("set env: %v\n", err)
		}

		var exVal bool
		if err := Bool(envName)(&exVal); err != nil {
			t.Fatalf("fnInt: %v\n", err)
		}

		assert.Equal(t, exVal, envVal)
	})

	t.Run("err parse", func(t *testing.T) {
		envVal := "x7"

		t.Cleanup(func() {
			os.Unsetenv(envName)
		})

		if err := os.Setenv(envName, envVal); err != nil {
			t.Fatalf("set env: %v\n", err)
		}

		var exVal bool
		if err := Bool(envName)(&exVal); err == nil {
			t.Fatal("want err")
		}
	})

	t.Run("err not set", func(t *testing.T) {
		var exVal bool

		err := Bool(envName)(&exVal)
		assert.Equal(t, err, perr.ErrNotSet)
	})
}

func TestString(t *testing.T) {
	envName := "myENV"

	t.Run("ok", func(t *testing.T) {
		envVal := "value"

		t.Cleanup(func() {
			os.Unsetenv(envName)
		})

		if err := os.Setenv(envName, envVal); err != nil {
			t.Fatalf("set env: %v\n", err)
		}

		var exVal string
		if err := String(envName)(&exVal); err != nil {
			t.Fatalf("fnInt: %v\n", err)
		}

		assert.Equal(t, exVal, envVal)
	})

	t.Run("err not set", func(t *testing.T) {
		var exVal string

		err := String(envName)(&exVal)
		assert.Equal(t, err, perr.ErrNotSet)
	})
}
