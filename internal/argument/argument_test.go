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
			"a",
			"A",
			"desc",
		),
		String(
			&tString,
			"b",
			"B",
			"desc",
		),
		Bool(
			&tBool,
			"c",
			"C",
			"desc",
		),
	)
	flag.Parse()

	for i := range args {
		args[i]()
	}
}
