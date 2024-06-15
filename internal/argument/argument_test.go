package argument

import (
	"flag"
	"testing"
)

func TestInt(t *testing.T) {
	tInt := int(10)
	tString := "def"
	tBool := true

	args := Array(
		Int(
			&tInt,
			"A",
		),
		String(
			&tString,
			"B",
		),
		Bool(
			&tBool,
			"C",
		),
	)
	flag.Parse()

	for i := range args {
		args[i]()
	}
}
