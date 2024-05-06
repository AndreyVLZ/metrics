package argument

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"strconv"
)

type Arg func() error

func Array(args ...Arg) []Arg { return args }

func Int(valInt *int, nameFlag string, nameENV string, usage string) Arg {
	return func() error {
		flag.IntVar(valInt, nameFlag, *valInt, usage)

		valIntParse, err := intEnv(nameENV)
		if err != nil {
			return err
		}

		*valInt = valIntParse

		return nil
	}
}

func String(valStr *string, nameFlag string, nameENV string, usage string) Arg {
	return func() error {
		flag.StringVar(valStr, nameFlag, *valStr, usage)

		valStrParse, err := stringEnv(nameENV)
		if err != nil {
			return err
		}

		*valStr = valStrParse

		return nil
	}
}

func Bool(valBool *bool, nameFlag string, nameENV string, usage string) Arg {
	return func() error {
		flag.BoolVar(valBool, nameFlag, *valBool, usage)

		valBoolParse, err := boolEnv(nameENV)
		if err != nil {
			return err
		}

		*valBool = valBoolParse

		return nil
	}
}

var (
	formatNotExist  = "does not exist env %s"
	formatIncorrect = "incorrect env %s"
)

type envError struct {
	errFormat string
}

func newErr(format string) *envError {
	return &envError{errFormat: format}
}

func (e *envError) Error(name string) string {
	return fmt.Sprintf(e.errFormat, name)
}

type IncorrectError struct {
	nameENV string
}

func (e IncorrectError) Error() string {
	return newErr(formatIncorrect).Error(e.nameENV)
}

type NotExistError struct {
	nameENV string
}

func (e NotExistError) Error() string {
	return newErr(formatNotExist).Error(e.nameENV)
}

func stringEnv(nameENV string) (string, error) {
	valStr, ok := os.LookupEnv(nameENV)
	if !ok {
		return "", NotExistError{nameENV: nameENV}
	}

	return valStr, nil
}

func intEnv(nameENV string) (int, error) {
	valStr, err := stringEnv(nameENV)
	if err != nil {
		return 0, err
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, IncorrectError{nameENV: nameENV}
	}

	return val, nil
}

func boolEnv(nameENV string) (bool, error) {
	valBool, ok := os.LookupEnv(nameENV)
	if !ok {
		return false, NotExistError{nameENV: nameENV}
	}

	return valBool == "true", nil
}
