package argument

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type arg func() error

func Array(args ...arg) []arg {
	return args
}

func Int(valInt *int, nameFlag string, nameENV string, usage string) arg {
	return func() error {
		flag.IntVar(valInt, nameFlag, *valInt, usage)
		valIntParse, err := intEnv(nameENV)
		if err != nil {
			return err
		}

		valInt = &valIntParse
		return nil
	}
}

func String(valStr *string, nameFlag string, nameENV string, usage string) arg {
	return func() error {
		flag.StringVar(valStr, nameFlag, *valStr, usage)
		valStrParse, err := stringEnv(nameENV)
		if err != nil {
			return err
		}

		valStr = &valStrParse
		return nil
	}
}

func Bool(valBool *bool, nameFlag string, nameENV string, usage string) arg {
	return func() error {
		flag.BoolVar(valBool, nameFlag, *valBool, usage)
		valBoolParse, err := boolEnv(nameENV)
		if err != nil {
			return err
		}

		valBool = &valBoolParse
		return nil
	}
}

var formatNotExist = "does not exist env %s"
var formatIncorrect = "incorrect env %s"

type errENV struct {
	errFormat string
}

func newErr(format string) *errENV {
	return &errENV{errFormat: format}
}

func (e *errENV) Error(name string) string {
	return fmt.Sprintf(e.errFormat, name)
}

type ErrIncorrect struct {
	nameENV string
}

func (e ErrIncorrect) Error() string {
	return newErr(formatIncorrect).Error(e.nameENV)
}

type ErrNotExist struct {
	nameENV string
}

func (e ErrNotExist) Error() string {
	return newErr(formatNotExist).Error(e.nameENV)
}

func stringEnv(nameENV string) (string, error) {
	valStr, ok := os.LookupEnv(nameENV)
	if !ok {
		return "", ErrNotExist{nameENV: nameENV}
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
		return 0, ErrIncorrect{nameENV: nameENV}
	}

	return val, nil
}

func boolEnv(nameENV string) (bool, error) {
	valBool, ok := os.LookupEnv(nameENV)
	if !ok {
		return false, ErrNotExist{nameENV: nameENV}
	}
	return valBool == "true", nil
}
