package flag

import (
	"bytes"
	"flag"
	"fmt"
	"testing"

	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		fs = flag.NewFlagSet("test", flag.ExitOnError)
		flagName := "a"

		fnInt := Int(flagName, "usage")
		exVal := 777

		if err := Parse([]string{fmt.Sprintf("-%s=%d", flagName, exVal)}); err != nil {
			t.Error(err)
		}

		var acVal int
		if err := fnInt(&acVal); err != nil {
			t.Error(err)
		}

		assert.Equal(t, exVal, acVal)
	})

	t.Run("err convert", func(t *testing.T) {
		fs = flag.NewFlagSet("test", flag.ContinueOnError)
		flagName := "a"
		var buf bytes.Buffer
		fs.SetOutput(&buf)

		Int(flagName, "usage")

		if err := Parse([]string{fmt.Sprintf("-%s=RT", flagName)}); err == nil {
			t.Error("want err")
		}
	})

	t.Run("err not set", func(t *testing.T) {
		fs = flag.NewFlagSet("test", flag.ExitOnError)
		flagName := "a"

		fnInt := Int(flagName, "usage")

		if err := Parse([]string{}); err != nil {
			t.Error(err)
		}

		var acVal int
		err := fnInt(&acVal)

		assert.Equal(t, perr.ErrNotSet, err)
	})
}

func TestBool(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		fs = flag.NewFlagSet("test", flag.ExitOnError)
		flagName := "a"

		fnBool := Bool(flagName, "usage")
		exVal := true

		if err := Parse([]string{fmt.Sprintf("-%s=%v", flagName, exVal)}); err != nil {
			t.Error(err)
		}

		var acVal bool
		if err := fnBool(&acVal); err != nil {
			t.Error(err)
		}

		assert.Equal(t, exVal, acVal)
	})

	t.Run("err convert", func(t *testing.T) {
		fs = flag.NewFlagSet("test", flag.ContinueOnError)
		flagName := "a"

		var buf bytes.Buffer
		fs.SetOutput(&buf)

		Bool(flagName, "usage")

		if err := Parse([]string{fmt.Sprintf("-%s=TT", flagName)}); err == nil {
			t.Error("want err")
		}
	})

	t.Run("err not set", func(t *testing.T) {
		fs = flag.NewFlagSet("test", flag.ExitOnError)
		flagName := "a"

		fnBool := Bool(flagName, "usage")

		if err := Parse([]string{}); err != nil {
			t.Error(err)
		}

		var acVal bool
		err := fnBool(&acVal)

		assert.Equal(t, perr.ErrNotSet, err)
	})
}

func TestString(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		fs = flag.NewFlagSet("test", flag.ExitOnError)
		flagName := "a"

		fnString := String(flagName, "usage")
		exVal := "myString"

		if err := Parse([]string{fmt.Sprintf("-%s=%s", flagName, exVal)}); err != nil {
			t.Error(err)
		}

		var acVal string
		if err := fnString(&acVal); err != nil {
			t.Error(err)
		}

		assert.Equal(t, exVal, acVal)
	})

	t.Run("err not set", func(t *testing.T) {
		fs = flag.NewFlagSet("test", flag.ExitOnError)
		flagName := "a"

		fnString := String(flagName, "usage")

		if err := Parse([]string{}); err != nil {
			t.Error(err)
		}

		var acVal string
		err := fnString(&acVal)

		assert.Equal(t, perr.ErrNotSet, err)
	})
}
