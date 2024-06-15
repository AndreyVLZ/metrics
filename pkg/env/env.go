package env

import (
	"fmt"
	"os"
	"strconv"
)

var (
	formatNotExist  = "does not exist env %s"
	formatIncorrect = "incorrect env %s"
)

type envError struct {
	errFormat string
	nameEnv   string
}

func (e *envError) Error() string {
	return fmt.Sprintf(e.errFormat, e.nameEnv)
}

func newIncorrectError(nameEnv string) error {
	return &envError{errFormat: formatIncorrect, nameEnv: nameEnv}
}

func newNotExistError(nameEnv string) error {
	return &envError{errFormat: formatNotExist, nameEnv: nameEnv}
}

type Arg func() error

func Array(args ...Arg) []Arg { return args }

// Возвращает функцию для парсинга Int env.
func Int(valInt *int, nameENV string) Arg {
	return func() error {
		valStr, isExist := stringEnv(nameENV)
		if !isExist {
			return newNotExistError(nameENV)
		}

		val, err := strconv.Atoi(valStr)
		if err != nil {
			return newIncorrectError(nameENV)
		}

		*valInt = val

		return nil
	}
}

// Парсинг String env.
func String(valStr *string, nameENV string) Arg {
	return func() error {
		valStrParse, isExist := stringEnv(nameENV)
		if !isExist {
			return newNotExistError(nameENV)
		}

		*valStr = valStrParse

		return nil
	}
}

func Bool(valBool *bool, nameENV string) Arg {
	return func() error {
		valBoolParse, isExist := boolEnv(nameENV)
		if !isExist {
			return newNotExistError(nameENV)
		}

		*valBool = valBoolParse

		return nil
	}
}

func stringEnv(nameENV string) (string, bool) {
	valStr, ok := os.LookupEnv(nameENV)
	if !ok {
		return "", false
	}

	return valStr, true
}

func boolEnv(nameENV string) (bool, bool) {
	valBool, ok := os.LookupEnv(nameENV)
	if !ok {
		return false, false
	}

	return valBool == "true", true
}
