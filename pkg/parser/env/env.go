package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
)

// Int Читает env как int. Возвращает функцию установки int-значения.
func Int(envName string) func(*int) error {
	return func(valInt *int) error {
		valStr, isExist := os.LookupEnv(envName)
		if !isExist {
			return perr.ErrNotSet
		}

		valInt1, err := strconv.Atoi(valStr)
		if err != nil {
			return fmt.Errorf("envInt conver [%s] to int: %w", valStr, err)
		}

		*valInt = valInt1

		return nil
	}
}

// Bool Читает env как bool. Возвращает функцию установки bool-значения.
func Bool(envName string) func(*bool) error {
	return func(valBool *bool) error {
		valStr, isExist := os.LookupEnv(envName)
		if !isExist {
			return perr.ErrNotSet
		}

		val, err := strconv.ParseBool(valStr)
		if err != nil {
			return fmt.Errorf("envBool convert [%s] to bool: %w", valStr, err)
		}

		*valBool = val

		return nil
	}
}

// String Читает env как string. Возвращает функцию установки string-значения.
func String(envName string) func(*string) error {
	return func(str *string) error {
		valStr, isExistENV := os.LookupEnv(envName)
		if !isExistENV {
			return perr.ErrNotSet
		}

		*str = valStr

		return nil
	}
}
